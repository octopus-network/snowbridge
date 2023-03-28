// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.19;

import "../../src/IOutboundQueue.sol";

contract OutboundQueueMock is IOutboundQueue {
    event Message(bytes dest, bytes payload);

    function submit(bytes calldata dest, bytes calldata payload) external payable {
        emit Message(dest, payload);
    }
}