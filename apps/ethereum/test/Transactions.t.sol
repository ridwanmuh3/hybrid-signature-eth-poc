// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.31;

import {Test} from "forge-std/Test.sol";
import {Transactions} from "../src/Transactions.sol";

contract TransactionsTest is Test {
    Transactions public transactions;
    address oracle = makeAddr("ORACLE");
    address sender = makeAddr("sender");
    address receiver = makeAddr("receiver");

    function setUp() public {
        transactions = new Transactions(oracle);
        vm.deal(sender, 10 ether);
        vm.prank(sender);
        transactions.deposit{value: 10 ether}();
    }

    function test_Constructor_SetsOracle() public view {
        assertEq(transactions.ORACLE(), oracle);
    }

    function test_Constructor_RejectsZeroOracle() public {
        vm.expectRevert("Invalid oracle address: zero address provided");
        new Transactions(address(0));
    }

    function test_Deposit_MetadataViews_DoNotRevert() public view {
        assertEq(transactions.balanceOf(sender), 10 ether);
        assertEq(transactions.name(), "Contract Ledger ETH");
        assertEq(transactions.symbol(), "ETH");
        assertEq(transactions.decimals(), 18);
        assertEq(transactions.tokenURI(0), "");
    }

    function test_SendTransaction_OnlyOracle() public {
        Transactions.TxParams memory params = _buildParams(1);
        vm.expectRevert("Only oracle can send transaction");
        transactions.sendTransaction(params);
    }

    function test_SendTransaction_RejectsZeroReceiver() public {
        Transactions.TxParams memory params = _buildParams(1);
        params.receiver = address(0);
        vm.prank(oracle);
        vm.expectRevert("Invalid receiver address");
        transactions.sendTransaction(params);
    }

    function test_SendTransaction_RejectsZeroSender() public {
        Transactions.TxParams memory params = _buildParams(1);
        params.sender = address(0);
        vm.prank(oracle);
        vm.expectRevert("Invalid sender address");
        transactions.sendTransaction(params);
    }

    function test_SendTransaction_RejectsSelfTransfer() public {
        Transactions.TxParams memory params = _buildParams(1);
        params.receiver = sender;
        vm.prank(oracle);
        vm.expectRevert("Sender equals receiver");
        transactions.sendTransaction(params);
    }

    function test_SendTransaction_RejectsZeroAmount() public {
        Transactions.TxParams memory params = _buildParams(1);
        params.amount = 0;
        vm.prank(oracle);
        vm.expectRevert("Amount must be non-zero");
        transactions.sendTransaction(params);
    }

    function test_SendTransaction_RejectsLowNonce() public {
        Transactions.TxParams memory params = _buildParams(0);
        vm.prank(oracle);
        vm.expectRevert("Nonce value too low");
        transactions.sendTransaction(params);
    }

    function test_SendTransaction_StoresAndEmitsEvent() public {
        Transactions.TxParams memory params = _buildParams(1);
        vm.prank(oracle);
        vm.expectEmit(true, true, false, true);
        emit Transactions.TransactionSent(1, sender, receiver, "ML-DSA-65", "single", "hello", 100);
        transactions.sendTransaction(params);

        assertEq(transactions.getUserNonce(sender), 1);
        Transactions.Transaction[] memory txs = transactions.getUserTransactions(sender);
        assertEq(txs.length, 1);
        assertEq(txs[0].nonce, 1);
        assertEq(txs[0].sender, sender);
        assertEq(txs[0].receiver, receiver);
        assertEq(txs[0].amount, 100);
    }

    function test_SendTransaction_NonceIncrement() public {
        Transactions.TxParams memory p1 = _buildParams(1);
        Transactions.TxParams memory p2 = _buildParams(2);
        vm.startPrank(oracle);
        transactions.sendTransaction(p1);
        transactions.sendTransaction(p2);
        vm.stopPrank();
        assertEq(transactions.getUserNonce(sender), 2);
        assertEq(transactions.getUserTransactions(sender).length, 2);
    }

    // Default helper: pure-PQC mode so on-chain ECDSA check is skipped (needsEcdsa=false).
    // Tests that exercise ECDSA explicitly override signingAlgorithm/signingMode and
    // supply a real 65-byte signature.
    function _buildParams(uint256 nonce) internal view returns (Transactions.TxParams memory) {
        return Transactions.TxParams({
            nonce: nonce,
            sender: sender,
            receiver: receiver,
            signingAlgorithm: "ML-DSA-65",
            signingMode: "single",
            message: "hello",
            amount: 100,
            ecdsaSignature: new bytes(0),
            pqSignature: new bytes(0)
        });
    }

    // =========================================================================
    // Security Tests
    // =========================================================================

    // -------------------------------------------------------------------------
    // Replay Attack — same nonce must be rejected after first use.
    // On-chain nonce is the last line of defence; oracle nonce check is the
    // first. This test proves the contract enforces monotonic nonces.
    // -------------------------------------------------------------------------

    function test_Security_RejectsNonceReplay() public {
        Transactions.TxParams memory params = _buildParams(1);
        vm.startPrank(oracle);
        transactions.sendTransaction(params);

        vm.expectRevert("Nonce value too low");
        transactions.sendTransaction(params); // same nonce → replay
        vm.stopPrank();
    }

    function test_Security_RejectsStaleNonce() public {
        // Advance to nonce=3 then try nonce=2 (old nonce).
        vm.startPrank(oracle);
        transactions.sendTransaction(_buildParams(1));
        transactions.sendTransaction(_buildParams(2));
        transactions.sendTransaction(_buildParams(3));

        vm.expectRevert("Nonce value too low");
        transactions.sendTransaction(_buildParams(2));
        vm.stopPrank();
    }

    // -------------------------------------------------------------------------
    // ECDSA Wrong-Sender — signature by a different key must be rejected.
    // Demonstrates that ecrecover binding to sender address is enforced.
    // -------------------------------------------------------------------------

    function test_Security_RejectsWrongSenderECDSA() public {
        uint256 attackerKey = 0xBAD;
        address attacker = vm.addr(attackerKey);

        // Attacker signs for their own address, but claims sender is the victim.
        bytes32 msgHash = _computeHash(1, attacker, receiver, "ECDSA", "single", "hello", 100);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(attackerKey, msgHash);
        bytes memory attackerSig = abi.encodePacked(r, s, v);

        Transactions.TxParams memory params = _buildParams(1);
        params.signingAlgorithm = "ECDSA";   // force needsEcdsa = true
        params.signingMode = "single";
        params.sender = sender;              // victim's address
        params.ecdsaSignature = attackerSig; // signed by attacker

        vm.prank(oracle);
        vm.expectRevert("On-chain ECDSA verification failed");
        transactions.sendTransaction(params);
    }

    // -------------------------------------------------------------------------
    // ECDSA Malleability — FIXED. The contract enforces low-s (EIP-2): for any
    // valid (r, s, v), the malleated counterpart (r, N-s, v') is rejected with
    // "Invalid s value". The canonical (low-s) signature still verifies.
    //
    // PQC components remain non-malleable by construction (Dilithium, Falcon,
    // MAYO, SNOVA do not have analogous malleability).
    // -------------------------------------------------------------------------

    uint256 constant SECP256K1_N =
        0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141;

    function test_Security_ECDSAMalleability_HighSRejected() public {
        uint256 signerKey = 0xA11CE;
        address signerAddr = vm.addr(signerKey);
        vm.deal(signerAddr, 1 ether);
        vm.prank(signerAddr);
        transactions.deposit{value: 1 ether}();

        vm.startPrank(oracle);

        // Canonical low-s signature: must succeed.
        transactions.sendTransaction(Transactions.TxParams({
            nonce: 1, sender: signerAddr, receiver: receiver,
            signingAlgorithm: "ECDSA", signingMode: "single", message: "hello", amount: 100,
            ecdsaSignature: _ecdsaSig(signerKey, 1, signerAddr), pqSignature: new bytes(0)
        }));

        // Malleated high-s twin: must revert with "Invalid s value".
        vm.expectRevert("Invalid s value");
        transactions.sendTransaction(Transactions.TxParams({
            nonce: 2, sender: signerAddr, receiver: receiver,
            signingAlgorithm: "ECDSA", signingMode: "single", message: "hello", amount: 100,
            ecdsaSignature: _malleableEcdsaSig(signerKey, 2, signerAddr), pqSignature: new bytes(0)
        }));
        vm.stopPrank();

        // Only the canonical sig was stored.
        assertEq(transactions.getUserTransactions(signerAddr).length, 1);
    }

    // -------------------------------------------------------------------------
    // PQ Signature Hash Binding — the contract stores keccak256(pqSignature).
    // A substituted PQ signature would produce a different stored hash, so
    // off-chain integrity checks can detect tampering.
    // -------------------------------------------------------------------------

    function test_Security_PQSignatureHashBinding() public {
        bytes memory pqSig      = hex"deadbeef01020304";
        bytes memory forgeSig   = hex"cafebabe01020304";

        Transactions.TxParams memory params = _buildParams(1);
        params.pqSignature = pqSig;
        vm.prank(oracle);
        transactions.sendTransaction(params);

        Transactions.Transaction[] memory txs = transactions.getUserTransactions(sender);
        bytes32 storedHash = txs[0].pqSignatureHash;

        assertEq(storedHash, keccak256(pqSig));
        assertTrue(storedHash != keccak256(forgeSig), "forged sig must produce different hash");
    }

    // -------------------------------------------------------------------------
    // Algorithm Field Binding — signingAlgorithm is part of the signed hash,
    // so an attacker cannot reuse a signature by relabelling the algorithm.
    // -------------------------------------------------------------------------

    function test_Security_AlgorithmFieldBinding() public {
        uint256 signerKey  = 0xC0FFEE;
        address signerAddr = vm.addr(signerKey);
        vm.deal(signerAddr, 1 ether);
        vm.prank(signerAddr);
        transactions.deposit{value: 1 ether}();

        // Sign with algorithm = "ECDSA".
        bytes32 msgHash = _computeHash(1, signerAddr, receiver, "ECDSA", "single", "hello", 100);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(signerKey, msgHash);
        bytes memory sig = abi.encodePacked(r, s, v);

        // Submit same sig but claim algorithm = "Hybrid-Secp256k1-ML-DSA-65".
        Transactions.TxParams memory params = Transactions.TxParams({
            nonce: 1,
            sender: signerAddr,
            receiver: receiver,
            signingAlgorithm: "Hybrid-Secp256k1-ML-DSA-65", // relabelled
            signingMode: "hybrid",
            message: "hello",
            amount: 100,
            ecdsaSignature: sig,
            pqSignature: new bytes(0)
        });

        vm.prank(oracle);
        vm.expectRevert("On-chain ECDSA verification failed");
        transactions.sendTransaction(params);
    }

    // -------------------------------------------------------------------------
    // Foundry Anvil account keys — mirrors what the Go oracle must use.
    // ECDSA part of any hybrid key for account N must be derived from key N.
    // -------------------------------------------------------------------------

    uint256 constant ANVIL_KEY_0 = 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80;
    uint256 constant ANVIL_KEY_9 = 0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6;

    // Replicates the exact failing RPC call.
    // Oracle = Anvil #9, sender = Anvil #0.
    // ECDSA sig must be signed by SENDER's key, not a random hybrid key.
    function test_Hybrid_AnvilKeys_ECDSASyncedWithSender() public {
        address anvilOracle = vm.addr(ANVIL_KEY_9);
        address anvilSender = vm.addr(ANVIL_KEY_0);

        // Deploy with Anvil oracle so prank below works.
        Transactions txs = new Transactions(anvilOracle);
        vm.deal(anvilSender, 10 ether);
        vm.prank(anvilSender);
        txs.deposit{value: 10 ether}();

        bytes32 msgHash = _computeHash(
            1,
            anvilSender,
            vm.addr(0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d), // Anvil #1
            "Hybrid-Secp256k1-ML-DSA-65",
            "hybrid",
            "Hello",
            5 ether
        );

        // Sign with SENDER's key — this is what the Go oracle must do when
        // building the hybrid private key (ecdsa_private_key = ANVIL_KEY_0).
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(ANVIL_KEY_0, msgHash);
        bytes memory ecdsaSig = abi.encodePacked(r, s, v);

        Transactions.TxParams memory params = Transactions.TxParams({
            nonce: 1,
            sender: anvilSender,
            receiver: vm.addr(0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d),
            signingAlgorithm: "Hybrid-Secp256k1-ML-DSA-65",
            signingMode: "hybrid",
            message: "Hello",
            amount: 5 ether,
            ecdsaSignature: ecdsaSig,
            pqSignature: new bytes(0)
        });

        vm.prank(anvilOracle);
        txs.sendTransaction(params);

        assertEq(txs.getUserNonce(anvilSender), 1);
    }

    // =========================================================================
    // Reentrancy Tests
    // =========================================================================
    //
    // ReentrancyAttacker: calls deposit then withdraw, and on receiving ETH
    // immediately tries to re-enter withdraw. Uses try-catch so receive() does
    // not revert (preventing the call from short-circuiting into "ETH transfer
    // failed"). The CEI pattern (balance zeroed BEFORE the external call) means
    // the re-entry finds balance == 0 and is blocked.

    function test_Security_Reentrancy_CEIBlocksAttacker() public {
        ReentrancyAttacker attacker = new ReentrancyAttacker(address(transactions));
        vm.deal(address(attacker), 2 ether);

        attacker.depositAndAttack{value: 2 ether}(2 ether, 2 ether);

        // receive() called once (legit withdrawal), re-entry caught and blocked
        assertEq(attacker.reentryCount(), 1);
        assertEq(attacker.stolenWei(), 2 ether); // only own deposit returned
        assertEq(transactions.balances(sender), 10 ether); // victim untouched
        assertEq(address(transactions).balance, 10 ether); // 12 - 2 = 10
    }

    // =========================================================================
    // Front-Running Tests
    // =========================================================================
    //
    // Front-running in this system: an attacker observes a pending oracle
    // sendTransaction call in the mempool and tries to submit it with modified
    // params (redirect receiver) or submit before the oracle.

    function test_Security_FrontRun_NonOracleCannotSubmitOracleTx() public {
        address frontrunner = makeAddr("frontrunner");
        Transactions.TxParams memory params = _buildParams(1);
        params.receiver = frontrunner; // try to redirect funds to self

        vm.prank(frontrunner);
        vm.expectRevert("Only oracle can send transaction");
        transactions.sendTransaction(params);
    }

    function test_Security_FrontRun_NoncePreventsDoubleExecution() public {
        // Nonce is the last line of defence against oracle replaying its own tx.
        Transactions.TxParams memory params = _buildParams(1);
        vm.startPrank(oracle);
        transactions.sendTransaction(params);

        vm.expectRevert("Nonce value too low");
        transactions.sendTransaction(params); // same nonce blocked
        vm.stopPrank();
    }

    // =========================================================================
    // Malicious Relayer Tests
    // =========================================================================

    // FIXED (was L1): when signingAlgorithm is "ECDSA" or signingMode is "hybrid",
    // the contract now requires exactly 65 bytes of ECDSA signature. A malicious
    // oracle can no longer redirect funds by submitting a non-65-byte ecdsaSignature.
    function test_Security_MaliciousRelayer_ShortSigRejected() public {
        address malicious = makeAddr("maliciousReceiver");
        Transactions.TxParams memory params = _buildParams(1);
        params.signingAlgorithm = "ECDSA";       // forces needsEcdsa
        params.signingMode = "single";
        params.receiver = malicious;
        params.ecdsaSignature = new bytes(64);   // not 65 → must revert

        vm.prank(oracle);
        vm.expectRevert("ECDSA signature required");
        transactions.sendTransaction(params);

        assertEq(transactions.balances(malicious), 0, "malicious receiver must not gain funds");
    }

    function test_Security_MaliciousRelayer_CannotInflateAmountBeyondBalance() public {
        Transactions.TxParams memory params = _buildParams(1);
        params.amount = 100 ether; // sender only has 10 ether

        vm.prank(oracle);
        vm.expectRevert("Insufficient contract balance");
        transactions.sendTransaction(params);
    }

    // Oracle tries to forge sender's ECDSA signature to redirect funds to itself.
    // ecrecover returns oracle's address, but params.sender = sender → mismatch.
    function test_Security_MaliciousRelayer_ForgedSignatureRejected() public {
        uint256 oracleKey = 0xBEEF;
        address oracleAddr = vm.addr(oracleKey);
        Transactions transactions2 = new Transactions(oracleAddr);
        vm.deal(sender, 5 ether);
        vm.prank(sender);
        transactions2.deposit{value: 5 ether}();

        bytes32 msgHash = _computeHash(1, sender, oracleAddr, "ECDSA", "single", "hello", 100);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(oracleKey, msgHash);
        bytes memory forgedSig = abi.encodePacked(r, s, v);

        Transactions.TxParams memory params = Transactions.TxParams({
            nonce: 1, sender: sender, receiver: oracleAddr,
            signingAlgorithm: "ECDSA", signingMode: "single", message: "hello", amount: 100,
            ecdsaSignature: forgedSig, pqSignature: new bytes(0)
        });

        vm.prank(oracleAddr);
        vm.expectRevert("On-chain ECDSA verification failed");
        transactions2.sendTransaction(params);
    }

    // =========================================================================
    // Signature Forgery Tests
    // =========================================================================

    // FIXED (was L1): when signingAlgorithm is "ECDSA" (or signingMode is "hybrid"),
    // empty ecdsaSignature is now rejected by the explicit length-65 require.
    function test_Security_Forgery_EmptySigRejected() public {
        Transactions.TxParams memory params = _buildParams(1);
        params.signingAlgorithm = "ECDSA"; // triggers needsEcdsa
        params.signingMode = "single";
        // ecdsaSignature stays bytes(0)

        vm.prank(oracle);
        vm.expectRevert("ECDSA signature required");
        transactions.sendTransaction(params);
        assertEq(transactions.getUserNonce(sender), 0, "nonce must not advance on revert");
    }

    function test_Security_Forgery_CorrectLengthWrongSigRejected() public {
        // sig.length == 65 + needsEcdsa activates ECDSA check; all-zero bytes
        // recover address(0) which never equals sender → reverts.
        Transactions.TxParams memory params = _buildParams(1);
        params.signingAlgorithm = "ECDSA";
        params.signingMode = "single";
        params.ecdsaSignature = new bytes(65); // all zeros, not a valid secp256k1 sig

        vm.prank(oracle);
        vm.expectRevert("On-chain ECDSA verification failed");
        transactions.sendTransaction(params);
    }

    function test_Security_Forgery_SigForDifferentAmountRejected() public {
        // Sig over (nonce=1, amount=100) cannot be reused for (nonce=1, amount=1 ether).
        uint256 signerKey = 0xDEAD;
        address signerAddr = vm.addr(signerKey);
        vm.deal(signerAddr, 5 ether);
        vm.prank(signerAddr);
        transactions.deposit{value: 5 ether}();

        bytes32 h = _computeHash(1, signerAddr, receiver, "ECDSA", "single", "hello", 100);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(signerKey, h);
        bytes memory sig = abi.encodePacked(r, s, v);

        Transactions.TxParams memory params = Transactions.TxParams({
            nonce: 1, sender: signerAddr, receiver: receiver,
            signingAlgorithm: "ECDSA", signingMode: "single", message: "hello",
            amount: 1 ether, // different from signed amount → different hash
            ecdsaSignature: sig, pqSignature: new bytes(0)
        });

        vm.prank(oracle);
        vm.expectRevert("On-chain ECDSA verification failed");
        transactions.sendTransaction(params);
    }

    // =========================================================================
    // Deposit Tests
    // =========================================================================

    function test_Deposit_Success() public {
        address depositor = makeAddr("depositor");
        vm.deal(depositor, 2 ether);
        vm.prank(depositor);
        vm.expectEmit(true, false, false, true);
        emit Transactions.Deposited(depositor, 2 ether);
        transactions.deposit{value: 2 ether}();
        assertEq(transactions.balances(depositor), 2 ether);
    }

    function test_Deposit_RejectsZeroValue() public {
        vm.prank(sender);
        vm.expectRevert("Must deposit non-zero ETH");
        transactions.deposit{value: 0}();
    }

    // =========================================================================
    // Withdraw Tests
    // =========================================================================

    function test_Withdraw_Success() public {
        uint256 prevSenderBalance = address(sender).balance;
        vm.prank(sender);
        vm.expectEmit(true, false, false, true);
        emit Transactions.Withdrawn(sender, 1 ether);
        transactions.withdraw(1 ether);
        assertEq(transactions.balances(sender), 9 ether);
        assertEq(address(sender).balance, prevSenderBalance + 1 ether);
    }

    function test_Withdraw_RejectsZeroAmount() public {
        vm.prank(sender);
        vm.expectRevert("Must withdraw non-zero ETH");
        transactions.withdraw(0);
    }

    function test_Withdraw_RejectsInsufficientBalance() public {
        vm.prank(sender);
        vm.expectRevert("Insufficient balance");
        transactions.withdraw(11 ether);
    }

    function test_Withdraw_CEI_NoReentrancy() public {
        assertEq(transactions.balances(sender), 10 ether);
        vm.prank(sender);
        transactions.withdraw(5 ether);
        assertEq(transactions.balances(sender), 5 ether);
        vm.prank(sender);
        transactions.withdraw(5 ether);
        assertEq(transactions.balances(sender), 0);
    }

    // -------------------------------------------------------------------------
    // Internal helper: replicates the contract's verifySignature hash
    // -------------------------------------------------------------------------

    function _ecdsaSig(uint256 key, uint256 nonce, address from) internal view returns (bytes memory) {
        bytes32 h = _computeHash(nonce, from, receiver, "ECDSA", "single", "hello", 100);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(key, h);
        return abi.encodePacked(r, s, v);
    }

    function _malleableEcdsaSig(uint256 key, uint256 nonce, address from) internal view returns (bytes memory) {
        bytes32 h = _computeHash(nonce, from, receiver, "ECDSA", "single", "hello", 100);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(key, h);
        return abi.encodePacked(r, bytes32(SECP256K1_N - uint256(s)), v == 27 ? uint8(28) : uint8(27));
    }

    function _computeHash(
        uint256 nonce,
        address _sender,
        address _receiver,
        string memory algo,
        string memory mode,
        string memory message,
        uint256 amount
    ) internal view returns (bytes32) {
        // Mirror Transactions.verifySignature exactly: chainid is the first
        // ABI-encoded element; omitting it would make every signed hash differ
        // from the on-chain reconstruction by 32 bytes and fail recovery.
        bytes32 messageHash = keccak256(
            abi.encode(block.chainid, nonce, _sender, _receiver, algo, mode, message, amount)
        );
        return keccak256(
            abi.encodePacked("\x19ETHEREUM-ORACLE-SIGNED:\n32", messageHash)
        );
    }
}

// ---------------------------------------------------------------------------
// Helper: reentrancy attacker. On receive, immediately tries to re-enter
// withdraw. Uses try-catch so receive() does not revert (which would mask the
// CEI protection as "ETH transfer failed" rather than demonstrating that the
// re-entry itself is blocked by a zero balance).
// ---------------------------------------------------------------------------
contract ReentrancyAttacker {
    Transactions private target;
    uint256 public reentryCount;
    uint256 public stolenWei;
    uint256 private _attackAmount;

    constructor(address _target) {
        target = Transactions(payable(_target));
    }

    function depositAndAttack(uint256 depositAmount, uint256 attackAmount) external payable {
        require(msg.value == depositAmount, "value mismatch");
        _attackAmount = attackAmount;
        if (depositAmount > 0) target.deposit{value: depositAmount}();
        if (attackAmount > 0) target.withdraw(attackAmount);
    }

    receive() external payable {
        reentryCount++;
        stolenWei += msg.value;
        if (_attackAmount > 0 && address(target).balance >= _attackAmount) {
            try target.withdraw(_attackAmount) {} catch {}
        }
    }
}
