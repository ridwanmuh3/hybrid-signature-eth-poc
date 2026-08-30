// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.31;

import {Script, console} from "forge-std/Script.sol";
import {Transactions} from "../src/Transactions.sol";

contract Deploy is Script {
    function run() external returns (Transactions) {
        address oracle = vm.envAddress("ORACLE_ADDRESS");

        vm.startBroadcast();
        Transactions transactions = new Transactions(oracle);
        vm.stopBroadcast();

        console.log("Transactions deployed at:", address(transactions));
        return transactions;
    }
}
