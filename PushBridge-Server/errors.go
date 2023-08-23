package main

import "errors"

var errPacketTooLarge = errors.New("packet Exceeds Max FCM Packet Size (4000 bytes)")
