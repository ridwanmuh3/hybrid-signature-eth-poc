// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.31;

contract Transactions {
    address public immutable ORACLE;
    mapping(address => uint) public balances;

    string internal constant LEDGER_SYMBOL = "ETH";
    string internal constant LEDGER_NAME = "Contract Ledger ETH";
    uint8 internal constant LEDGER_DECIMALS = 18;

    event TransactionSent(
        uint256 nonce,
        address indexed sender,
        address indexed receiver,
        string signingAlgorithm,
        string signingMode,
        string message,
        uint256 amount
    );

    struct Transaction {
        uint256 nonce;
        address sender;
        address receiver;
        string signingAlgorithm;
        string signingMode;
        string message;
        uint256 amount;
        uint256 timestamp;
        bytes ecdsaSignature;
        bytes32 pqSignatureHash;
    }

    struct TxParams {
        uint256 nonce;
        address sender;
        address receiver;
        string signingAlgorithm;
        string signingMode;
        string message;
        uint256 amount;
        bytes ecdsaSignature;
        bytes pqSignature;
    }

    mapping(address => Transaction[]) public userTransactions;
    mapping(address => uint256) public userNonce;

    event Deposited(address indexed user, uint256 amount);
    event Withdrawn(address indexed user, uint256 amount);

    constructor(address _oracle) {
        require(
            _oracle != address(0),
            "Invalid oracle address: zero address provided"
        );
        ORACLE = _oracle;
    }

    function deposit() public payable {
        require(msg.value > 0, "Must deposit non-zero ETH");
        balances[msg.sender] += msg.value;
        emit Deposited(msg.sender, msg.value);
    }

    // Wallet/RPC simulation paths sometimes probe touched contracts like ERC-20s.
    // Mirror standard metadata so those probes do not revert on deposit/withdraw.
    function balanceOf(address account) public view returns (uint256) {
        return balances[account];
    }

    function name() public pure returns (string memory) {
        return LEDGER_NAME;
    }

    function symbol() public pure returns (string memory) {
        return LEDGER_SYMBOL;
    }

    function decimals() public pure returns (uint8) {
        return LEDGER_DECIMALS;
    }

    function tokenURI(uint256) public pure returns (string memory) {
        return "";
    }

    function withdraw(uint256 amount) public {
        require(amount > 0, "Must withdraw non-zero ETH");
        require(balances[msg.sender] >= amount, "Insufficient balance");

        balances[msg.sender] -= amount;

        (bool success, ) = payable(msg.sender).call{value: amount}("");
        require(success, "ETH transfer failed");

        emit Withdrawn(msg.sender, amount);
    }

    function sendTransaction(TxParams memory _params) public {
        require(msg.sender == ORACLE, "Only oracle can send transaction");
        require(_params.sender != address(0), "Invalid sender address");
        require(_params.receiver != address(0), "Invalid receiver address");
        require(_params.sender != _params.receiver, "Sender equals receiver");
        require(_params.amount > 0, "Amount must be non-zero");
        require(
            _params.nonce > userNonce[_params.sender],
            "Nonce value too low"
        );
        require(balances[_params.sender] >= _params.amount, "Insufficient contract balance");

        bool needsEcdsa = keccak256(bytes(_params.signingMode)) == keccak256(bytes("hybrid")) ||
                          keccak256(bytes(_params.signingAlgorithm)) == keccak256(bytes("ECDSA"));
        if (needsEcdsa) {
            require(_params.ecdsaSignature.length == 65, "ECDSA signature required");
            require(verifySignature(_params), "On-chain ECDSA verification failed");
        }

        balances[_params.sender] -= _params.amount;
        balances[_params.receiver] += _params.amount;

        userNonce[_params.sender] = _params.nonce;
        bytes32 pgSignatureHash = keccak256(_params.pqSignature);


        userTransactions[_params.sender].push(
            Transaction({
                nonce: _params.nonce,
                sender: _params.sender,
                receiver: _params.receiver,
                signingAlgorithm: _params.signingAlgorithm,
                signingMode: _params.signingMode,
                message: _params.message,
                amount: _params.amount,
                timestamp: block.timestamp,
                ecdsaSignature: _params.ecdsaSignature,
                pqSignatureHash: pgSignatureHash
            })
        );

        emit TransactionSent(
            _params.nonce,
            _params.sender,
            _params.receiver,
            _params.signingAlgorithm,
            _params.signingMode,
            _params.message,
            _params.amount
        );
    }

    function verifySignature(
        TxParams memory _params
    ) internal view returns (bool) {
        bytes32 messageHash = keccak256(
            abi.encode(
                block.chainid,
                _params.nonce,
                _params.sender,
                _params.receiver,
                _params.signingAlgorithm,
                _params.signingMode,
                _params.message,
                _params.amount
            )
        );

        bytes32 ethSignedMessageHash = keccak256(
            abi.encodePacked("\x19ETHEREUM-ORACLE-SIGNED:\n32", messageHash)
        );

        return
            validateSigner(
                ethSignedMessageHash,
                _params.ecdsaSignature,
                _params.sender
            );
    }

    // secp256k1 group order halved — used for low-s enforcement (EIP-2 / SEC-1).
    // Rejecting high-s removes ECDSA signature malleability so that for any
    // signed hash there is exactly one canonical signature.
    uint256 internal constant SECP256K1_HALF_N =
        0x7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF5D576E7357A4501DDFE92F46681B20A0;

    function recoverSigner(
        bytes32 _ethSignedMessageHash,
        bytes memory _ecdsaSignature
    ) internal pure returns (address) {
        require(_ecdsaSignature.length == 65, "Invalid signature length");

        bytes32 r;
        bytes32 s;
        uint8 v;

        assembly {
            r := mload(add(_ecdsaSignature, 32))
            s := mload(add(_ecdsaSignature, 64))
            v := byte(0, mload(add(_ecdsaSignature, 96)))
        }

        // Reject high-s (EIP-2): only canonical signatures accepted.
        require(uint256(s) <= SECP256K1_HALF_N, "Invalid s value");

        if (v < 27) {
            v += 27;
        }
        require(v == 27 || v == 28, "Invalid v value");

        return ecrecover(_ethSignedMessageHash, v, r, s);
    }

    function getUserTransactions(
        address _user
    ) public view returns (Transaction[] memory) {
        return userTransactions[_user];
    }

    function getUserNonce(address _user) public view returns (uint256) {
        return userNonce[_user];
    }

    function validateSigner(
        bytes32 _messageHash,
        bytes memory _ecdsaSignature,
        address _expectedAddress
    ) internal pure returns (bool) {
        address signer = recoverSigner(_messageHash, _ecdsaSignature);
        return signer != address(0) && signer == _expectedAddress;
    }
}
