// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.8;

import "suave-std/suavelib/Suave.sol";

contract Passthrough {
    // These should just be hardcoded into the precompile call
    string public constant composerURL = "localhost:8080";
    string public constant composerEndpoint = "rollupBundle";

    event NewBundle(address ccrSender, bytes bundleBytes);

    function makeBundle() external payable {
        // Retrieve the rollup tx data from the confidential inputs
        bytes memory rollupTx = Suave.confidentialInputs();

        // Send POST request to the composer
        Suave.submitBundleJsonRPC(composerURL, composerEndpoint, rollupTx);
        emit NewBundle(msg.sender, rollupTx);
    }
}
