package protocol

import (
	"encoding/json"
	"errors"
	"github.com/hlhv/fsock"
)

/* FrameKind determines what a Frame interface should be cast to.
 */
type FrameKind byte

const (
	/* these numbers are part of the protocol and incredibly importatnt. Do
	 * not make this an enum with iota.
	 */
	// authentication/setup
	FrameKindIAm    FrameKind = 0x00
	FrameKindAccept FrameKind = 0x08

	// mounting
	FrameKindMount   FrameKind = 0x10
	FrameKindUnmount FrameKind = 0x11

	// resource requesting
	FrameKindNeedBand FrameKind = 0x20

	// http
	FrameKindHTTPReqHead FrameKind = 0x30
	FrameKindHTTPReqBody FrameKind = 0x31
	FrameKindHTTPReqEnd  FrameKind = 0x32

	FrameKindHTTPResWant FrameKind = 0x38
	FrameKindHTTPResHead FrameKind = 0x39
	FrameKindHTTPResBody FrameKind = 0x3A
	FrameKindHTTPResEnd  FrameKind = 0x3B
)

/* A connection can be for a cell, or a band. These constants store the
 * connection kind numbers.
 */
const (
	/* these numbers are part of the protocol and incredibly importatnt. Do
	 * not make this an enum with iota.
	 */
	ConnKindCell = 0x0
	ConnKindBand = 0x1
)

/* Frame represents a singular block of data sent between cells.
 */
type Frame interface {
	Kind() FrameKind
}

/* DataFrame is a specific type of frame which can store arbitrary data.
 */
type DataFrame interface {
	GetData() []byte
}

/* FrameIAm is sent from the client cell to the queen in order to initiate a
 * connection.
 */
type FrameIAm struct {
	ConnKind int    `json:"connKind"`
	Uuid     string `json:"uuid"`
	Key      string `json:"key"`
}

/* FrameAccept is sent from the queen to the client cell, assigning the cell a
 * UUID and session key that it can use to create bands later.
 */
type FrameAccept struct {
	Uuid string `json:"uuid"`
	Key  string `json:"key"`
}

/* FrameMount is sent from the client cell to the queen. It contains information
 * about the location the cell wants to mount on.
 */
type FrameMount struct {
	Host string `json:"host"`
	Path string `json:"path"`
}

/* FrameUnmount is sent from the client cell to the queen. Upon receiving this,
 * the queen will unmount the cell from its current mount point.
 */
type FrameUnmount struct{}

/* FrameNeedBand is sent from the queen to the client cell when the queen
 * detects that more bands are needed to keep the connection running smoothly.
 * The cell is expected to immediately create and connect a new band to the
 * queen.
 */
type FrameNeedBand struct {
	Count int `json:"count"`
}

/* FrameHTTPReqHead is sent from the queen to the client cell when the queen
 * receives an HTTP request and has determined that this cell should handle it.
 * It contains extensive information about the request. If the cell wants the
 * HTTP body to be sent to it, it must specifically request it.
 */
type FrameHTTPReqHead struct {
	RemoteAddrReal string `json:"remoteAddrReal"`
	RemoteAddr     string `json:"remoteAddr"`
	Method         string `json:"method"`

	Scheme   string              `json:"scheme"`
	Host     string              `json:"host"`
	Port     int                 `json:"port"`
	Path     string              `json:"path"`
	Fragment string              `json:"fragment"`
	Query    map[string][]string `json:"query"`

	Proto      string `json:"proto"`
	ProtoMajor int    `json:"protoMajor"`
	ProtoMinor int    `json:"protoMinor"`

	Headers map[string][]string `json:"headers"`
}

/* FrameHTTPReqBody is sent from the queen to the client cell after the cell
 * asks for the HTTP body. It contains a single chunk of the body, and is often
 * sent mulitple times in a row. The cell should stitch these together until
 * it receives FrameHTTPReqEnd.
 */
type FrameHTTPReqBody struct {
	Data []byte
}

/* FrameHTTPReqEnd is sent from the queen to the client cell when there is no
 * more HTTP body data left.
 */
type FrameHTTPReqEnd struct{}

/* FrameHTTPResWant is sent from the client cell to the queen as a request for
 * the HTTP request body data. The cell must specify a maximum size. The queen,
 * upon receiving this, will begin sending the HTTP body in chunks.
 */
type FrameHTTPResWant struct {
	MaxSize int `json:"maxSize"`
}

/* FrameHTTPResHead is sent from the client cell to the queen. It contains
 * information such as the status code, and headers. It signals to the queen
 * that it should begin responding to the HTTP client.
 */
type FrameHTTPResHead struct {
	StatusCode int                 `json:"statusCode"`
	Headers    map[string][]string `json:"headers"`
}

/* FrameHTTPResBody is sent from the client cell to the queen. It contains a
 * chunk of the HTTP response body data that should be written to the client.
 * The queen, upon receiving this, should send it to the HTTP client. This frame
 * should not precede FrameHTTPResHead.
 */
type FrameHTTPResBody struct {
	Data []byte
}

/* FrameHTTPResEnd is sent from the client cell to the queen as a signal that
 * the entirety of the HTTP response body has been sent, and the connection to
 * the HTTP client should be closed.
 */
type FrameHTTPResEnd struct{}

/* These functions return the frame kind from frame structs so that it can be
 * determined which frame struct to cast a Frame interface to.
 */
func (frame *FrameIAm) Kind() FrameKind    { return FrameKindIAm }
func (frame *FrameAccept) Kind() FrameKind { return FrameKindAccept }

func (frame *FrameMount) Kind() FrameKind   { return FrameKindMount }
func (frame *FrameUnmount) Kind() FrameKind { return FrameKindUnmount }

func (frame *FrameNeedBand) Kind() FrameKind { return FrameKindNeedBand }

func (frame *FrameHTTPReqHead) Kind() FrameKind { return FrameKindHTTPReqHead }
func (frame *FrameHTTPReqBody) Kind() FrameKind { return FrameKindHTTPReqBody }
func (frame *FrameHTTPReqEnd) Kind() FrameKind  { return FrameKindHTTPReqEnd }

func (frame *FrameHTTPResWant) Kind() FrameKind { return FrameKindHTTPResWant }
func (frame *FrameHTTPResHead) Kind() FrameKind { return FrameKindHTTPResHead }
func (frame *FrameHTTPResBody) Kind() FrameKind { return FrameKindHTTPResBody }
func (frame *FrameHTTPResEnd) Kind() FrameKind  { return FrameKindHTTPResEnd }

/* These functions return the arbitrary data stored in data frames.
 */
func (frame *FrameHTTPReqBody) GetData() []byte { return frame.Data }
func (frame *FrameHTTPResBody) GetData() []byte { return frame.Data }

/* ParseFrame splits a frame into its kind and its data. Unmarshaling should be
 * conducted by the handler.
 */
func ParseFrame(frameData []byte) (kind FrameKind, data []byte, err error) {
	if len(frameData) < 1 {
		return 0, nil, errors.New("empty frame")
	}
	return FrameKind(frameData[0]), frameData[1:], nil
}

/* ReadParseFrame reads a frame from an fsock reader and parses it.
 */
func ReadParseFrame(
	reader *fsock.Reader,
) (
	kind FrameKind,
	data []byte,
	err error,
) {
	frame, err := reader.Read()
	if err != nil {
		return 0, nil, err
	}
	return ParseFrame(frame)
}

/* MarshalFrame takes in a struct satisfying the Frame interface and endcodes it
 * into a valid frame.
 */
func MarshalFrame(
	frame Frame,
) (
	frameData []byte,
	err error,
) {
	var data []byte

	switch frame.Kind() {
	// some types have a []byte payload
	case FrameKindHTTPReqBody:
	case FrameKindHTTPResBody:
		data = frame.(DataFrame).GetData()
		break

	// some types have no payload at all
	case FrameKindHTTPReqEnd:
	case FrameKindHTTPResEnd:
		break

	// ... but most types will have a json payload (later will be hnbs)
	default:
		data, err = json.Marshal(frame)
		if err != nil {
			return nil, err
		}
		break
	}

	frameData = append(data, 0)
	copy(frameData[1:], frameData)
	frameData[0] = byte(frame.Kind())
	return
}

/* WriteMarshalFrame marshals and writes a Frame.
 */
func WriteMarshalFrame(writer *fsock.Writer, frame Frame) (nn int, err error) {
	frameData, err := MarshalFrame(frame)
	if err != nil {
		return 0, err
	}
	return writer.WriteFrame(frameData)
}
