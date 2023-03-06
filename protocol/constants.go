package protocol

const PROTOCOL_HEADER uint32 = 4
const FRAGMENT_HEADER uint32 = 4
const MESSAGE_MAX_SIZE uint32 = 4096
const MESSAGE_LENGTH uint32 = PROTOCOL_HEADER + MESSAGE_MAX_SIZE + 1