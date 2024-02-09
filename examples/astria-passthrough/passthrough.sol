// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.8;

import "suave-std/suavelib/Suave.sol";

contract Passthrough {
    // These should just be hardcoded into the precompile call
    string public constant composerURL = "http://localhost:8080/rollupBundle";
    string public constant composerEndpoint = "rollupBundle";

    event NewBundle(address ccrSender, bytes bundleBytes);

    function makeBundle() external payable {
        // Retrieve the rollup tx data from the confidential inputs
        bytes memory rollupTx = Suave.confidentialInputs();

        // Send POST request to the composer
        Suave.HttpRequest memory request = Suave.HttpRequest({
            url: composerURL,
            method: "POST",
            headers: new string[](0),
            body: rollupTx,
            withFlashbotsSignature: true
        });
        Suave.doHTTPRequest(request);
        emit NewBundle(msg.sender, rollupTx);
    }
}
