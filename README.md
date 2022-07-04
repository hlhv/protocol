# protocol

This module contains functionality to implement the current HLHV cell-to-cell
communication protocol.

## Protocol Overview

The HLHV protocol is based on a client/server model, where a central server
called the queen cell routes incoming HTTPS requests to different client cells.
Client cells connect to the queen cell over a main socket called the leash, and
maintain a second multiplexed connection made of several other sockets called
bands. Bands are used for shoveling HTTP requests and responses back and forth
at an alarming rate, and they are automatically created and discarded as needed.
The leash is for sending messages that control and regulate the overall
connection, such as mounting, un-mounting, and requesting more bands.

```
┌────────┐                   ┌─────────┐                   ┌────────┐
│        │ ◄──── Leash ────► │         │ ◄──── Leash ────► │        │
│        │                   │         │                   │        │
│  Cell  │ ◄───────────────► │  Queen  │ ◄───────────────► │  Cell  │
│        │ ◄──── Bands ────► │         │ ◄──── Bands ────► │        │
│        │ ◄───────────────► │         │ ◄───────────────► │        │
└────────┘                   └─────────┘                   └────────┘
```

## Connection Initiation

All frames mentioned in this section are sent over the leash.

The queen listens for incoming connections on what is usually port 2001. Cells
connect to this port over TLS, using the [fsock](https://github.com/hlhv/fsock)
socket framing protocol. The queen and client cell communicate by sending JSON
encoded frames (at a later date, this will be replaced with HNBS), and frames
containing arbitrary binary data.

To initiate the connection, the client cell must send a frame of type FrameIAm.
This frame contains the connection kind (which must be ConnKindCell, aka 0x0),
and the correct key.

If the key is correct, the queen will send a frame of type FrameAccept,
containing the connection UUID, and the connection key. These will be used later
by the cell in order to connect new bands. If the key is incorrect, or a timeout
was reached, the queen should close the connection.

Before HTTP requests can be sent to the cell, it must first request to mount on
the queen. To do this, it must send a frame of type FrameMount with the hostname
and path it wants to mount on. If the path ends with a '/', the cell will
receive requests for all paths under the path it has mounted on (except for
paths that other cells are mounted on).

To un-mount, the client must send a frame of type FrameUnmount. This frame
carries no data.
