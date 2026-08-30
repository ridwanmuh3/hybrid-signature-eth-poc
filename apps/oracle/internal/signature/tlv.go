package signature

// Hybrid blob wire format:
//
//	[MAGIC 4B big-endian] [VERSION 1B] [FIELD ...]
//	FIELD = [TAG 1B] [LENGTH 4B big-endian] [VALUE LENGTH×B]
//
// Magic = 0x48594253 ("HYBS")
// Version = 0x01
//
// Tags:
//
//	0x01  algorithm name (UTF-8 string)
//	0x02  ECDSA private key (32 B secp256k1 scalar)
//	0x03  ECDSA public key  (65 B uncompressed)
//	0x04  PQC private key   (liboqs raw)
//	0x05  PQC public key    (liboqs raw)
//	0x06  ECDSA signature   (65 B)
//	0x07  PQC signature     (liboqs raw)
//
// The algorithm tag is mandatory in every blob so that the type of the
// enclosing blob is self-describing and cannot be silently mis-routed.

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	tlvMagic   = uint32(0x48594253) // "HYBS"
	tlvVersion = byte(0x01)

	TagAlgorithm byte = 0x01
	TagECDSAPriv byte = 0x02
	TagECDSAPub  byte = 0x03
	TagPQCPriv   byte = 0x04
	TagPQCPub    byte = 0x05
	TagECDSASig  byte = 0x06
	TagPQCSig    byte = 0x07
)

var (
	ErrInvalidMagic   = errors.New("invalid TLV magic: not a hybrid blob")
	ErrInvalidVersion = errors.New("unsupported TLV version")
	ErrMissingField   = errors.New("required TLV field missing")
	ErrTruncated      = errors.New("TLV blob truncated")
	ErrDuplicateTag   = errors.New("duplicate TLV tag")
)

// tlvMaxBlobSize bounds an entire decoded TLV blob. liboqs's largest signature
// at our supported levels (Falcon-1024 ≤ 1280 B, MAYO-5 ≈ 838 B, ML-DSA-87
// 4 627 B, SNOVA_60_10_4 ≈ 576 B) plus public-key + algorithm string fits
// comfortably under 64 KiB. Anything larger is treated as adversarial input.
const tlvMaxBlobSize = 64 * 1024

type tlvField struct {
	tag   byte
	value []byte
}

// EncodeSigTLV encodes a hybrid signature into the TLV wire format.
func EncodeSigTLV(algorithm string, ecdsaSig, pqcSig []byte) []byte {
	return encodeTLV([]tlvField{
		{TagAlgorithm, []byte(algorithm)},
		{TagECDSASig, ecdsaSig},
		{TagPQCSig, pqcSig},
	})
}

// EncodePubKeyTLV encodes a hybrid public key into the TLV wire format.
func EncodePubKeyTLV(algorithm string, ecdsaPub, pqcPub []byte) []byte {
	return encodeTLV([]tlvField{
		{TagAlgorithm, []byte(algorithm)},
		{TagECDSAPub, ecdsaPub},
		{TagPQCPub, pqcPub},
	})
}

// EncodePrivKeyTLV encodes a hybrid private key into the TLV wire format.
func EncodePrivKeyTLV(algorithm string, ecdsaPriv, pqcPriv []byte) []byte {
	return encodeTLV([]tlvField{
		{TagAlgorithm, []byte(algorithm)},
		{TagECDSAPriv, ecdsaPriv},
		{TagPQCPriv, pqcPriv},
	})
}

// DecodeTLV parses a hybrid TLV blob and returns fields keyed by tag.
// Exported so tests and external verifiers can inspect individual components.
func DecodeTLV(data []byte) (map[byte][]byte, error) {
	return decodeTLV(data)
}

// encodeTLV serializes fields into the TLV wire format.
func encodeTLV(fields []tlvField) []byte {
	size := 5 // 4 magic + 1 version
	for _, f := range fields {
		size += 1 + 4 + len(f.value)
	}
	buf := make([]byte, 0, size)

	var magic [4]byte
	binary.BigEndian.PutUint32(magic[:], tlvMagic)
	buf = append(buf, magic[:]...)
	buf = append(buf, tlvVersion)

	for _, f := range fields {
		var lenBuf [4]byte
		binary.BigEndian.PutUint32(lenBuf[:], uint32(len(f.value)))
		buf = append(buf, f.tag)
		buf = append(buf, lenBuf[:]...)
		buf = append(buf, f.value...)
	}
	return buf
}

// decodeTLV validates magic + version then reads all fields into a map.
//
// Hardening properties:
//   - Rejects blobs larger than tlvMaxBlobSize to bound memory.
//   - Rejects duplicate tags so an attacker cannot smuggle a second value
//     for a critical tag (e.g. two TagECDSAPriv entries) and have the
//     last-writer-wins decoder pick a hostile value.
//   - All length arithmetic uses uint64 conversions so a maliciously large
//     declared length cannot overflow int and bypass the bounds check.
func decodeTLV(data []byte) (map[byte][]byte, error) {
	if len(data) > tlvMaxBlobSize {
		return nil, fmt.Errorf("%w: blob exceeds %d bytes", ErrTruncated, tlvMaxBlobSize)
	}
	if len(data) < 5 {
		return nil, ErrTruncated
	}
	if binary.BigEndian.Uint32(data[:4]) != tlvMagic {
		return nil, ErrInvalidMagic
	}
	if data[4] != tlvVersion {
		return nil, fmt.Errorf("%w: got 0x%02x", ErrInvalidVersion, data[4])
	}

	fields := make(map[byte][]byte)
	pos := uint64(5)
	end := uint64(len(data))
	for pos < end {
		if pos+5 > end {
			return nil, ErrTruncated
		}
		tag := data[pos]
		fieldLen := uint64(binary.BigEndian.Uint32(data[pos+1 : pos+5]))
		pos += 5
		if pos+fieldLen > end {
			return nil, fmt.Errorf("%w: field 0x%02x", ErrTruncated, tag)
		}
		if _, exists := fields[tag]; exists {
			return nil, fmt.Errorf("%w: tag 0x%02x", ErrDuplicateTag, tag)
		}
		v := make([]byte, fieldLen)
		copy(v, data[pos:pos+fieldLen])
		fields[tag] = v
		pos += fieldLen
	}
	return fields, nil
}

// requireField looks up a tag or returns ErrMissingField.
func requireField(fields map[byte][]byte, tag byte) ([]byte, error) {
	v, ok := fields[tag]
	if !ok {
		return nil, fmt.Errorf("%w: tag 0x%02x", ErrMissingField, tag)
	}
	return v, nil
}
