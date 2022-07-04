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

## HTTPS Requests

All frames mentioned in this section are sent over bands.

Upon receiving an HTTPS request, the queen directs information about the request
head to the proper cell in the form of a frame of type FrameHTTPReqHead.

This frame contains no request body data. If the client wants this data, it must
ask for it by sending a frame of type FrameHTTPResWant, specifying the maximum
length of the request body. The queen will then send the request body data in
chunks via frames of type FrameHTTPReqBody, ending with a frame of type
FrameHTTPReqHead. The cell is expected to stitch all body data frames together
in order in which they are received.

Then, the client must send a frame of type FrameHTTPResHead, containing the HTTP
status code of the response and a map containing headers. After this, the client
must send the response body in chunks via frames of type FrameHTTPResBody,
ending with a frame of type FrameHTTPResEnd.

## Band Requests

When there are no available bands to handle an HTTP request, the queen will
request that the client connect another band by sending a frame of type
FrameNeedBand over the leash. This frame will contain the number of new bands to
create. It is important that the client immediately respond to this request,
because the queen waits until the band has connected to resume fulfilling the
HTTP request.

The client should connect new bands in the same way as connecting as a cell,
but by specifying a connection kind of ConnKindBand (aka 0x1), and specifying
the connection UUID and connection key.
