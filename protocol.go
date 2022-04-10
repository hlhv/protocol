package protocol

import (
        "errors"
        "encoding/json"
        "github.com/hlhv/fsock"
)

type FrameKind byte

const (
        /* these numbers are part of the protocol and incredibly importatnt. Do
         * not make this an enum with iota.
         */
        // authentication/setup
        FrameKindIAm         FrameKind = 0x00
        FrameKindAccept      FrameKind = 0x08

        // mounting
        FrameKindMount       FrameKind = 0x10
        FrameKindUnmount     FrameKind = 0x11

        // resource requesting
        FrameKindNeedBand    FrameKind = 0x20

        // http
        FrameKindHTTPReqHead FrameKind = 0x30
        FrameKindHTTPReqBody FrameKind = 0x31
        FrameKindHTTPReqEnd  FrameKind = 0x32

        FrameKindHTTPResWant FrameKind = 0x38
        FrameKindHTTPResHead FrameKind = 0x39
        FrameKindHTTPResBody FrameKind = 0x3A
        FrameKindHTTPResEnd  FrameKind = 0x3B
)

const (
        /* these numbers are part of the protocol and incredibly importatnt. Do
         * not make this an enum with iota.
         */
        ConnKindCell = 0x0
        ConnKindBand = 0x1
)

type Frame interface {
        Kind() FrameKind
}

type DataFrame interface {
        GetData() []byte
}

// frame kind structs
type FrameIAm struct {
        ConnKind int    `json:"connKind"`
        Uuid     string `json:"uuid"`
        Key      string `json:"key"`
}

type FrameAccept struct {
        Uuid string `json:"uuid"`
        Key  string `json:"key"`
}

type FrameMount struct {
        Host string `json:"host"`
        Path string `json:"path"`
}

type FrameUnmount struct { }

type FrameNeedBand struct {
        Count int `json:"count"`
}

type FrameHTTPReqHead struct {
        RemoteAddrReal string               `json:"remoteAddrReal"`
        RemoteAddr     string               `json:"remoteAddr"`
        Method         string               `json:"method"`

        Scheme         string               `json:"scheme"`
        Host           string               `json:"host"`
        Port           int                  `json:"port"`
        Path           string               `json:"path"`
        Fragment       string               `json:"fragment"`
        Query          map[string] []string `json:"query"`

        Proto          string               `json:"proto"`
        ProtoMajor     int                  `json:"protoMajor"`
        ProtoMinor     int                  `json:"protoMinor"`

        Headers        map[string] []string `json:"headers"`
}

type FrameHTTPReqBody struct {
        Data []byte
}

type FrameHTTPReqEnd struct {}

type FrameHTTPResWant struct {
        MaxSize int `json:"maxSize"`
}

type FrameHTTPResHead struct {
        StatusCode int                  `json:"statusCode"`
        Headers    map[string] []string `json:"headers"`
}

type FrameHTTPResBody struct {
        Data []byte
}

type FrameHTTPResEnd struct {}

// frame kind getters
func (frame *FrameIAm)         Kind () FrameKind { return FrameKindIAm    }
func (frame *FrameAccept)      Kind () FrameKind { return FrameKindAccept }

func (frame *FrameMount)       Kind () FrameKind { return FrameKindMount   }
func (frame *FrameUnmount)     Kind () FrameKind { return FrameKindUnmount }

func (frame *FrameNeedBand)    Kind () FrameKind { return FrameKindNeedBand }

func (frame *FrameHTTPReqHead) Kind () FrameKind { return FrameKindHTTPReqHead }
func (frame *FrameHTTPReqBody) Kind () FrameKind { return FrameKindHTTPReqBody }
func (frame *FrameHTTPReqEnd)  Kind () FrameKind { return FrameKindHTTPReqEnd  }

func (frame *FrameHTTPResWant) Kind () FrameKind { return FrameKindHTTPResWant }
func (frame *FrameHTTPResHead) Kind () FrameKind { return FrameKindHTTPResHead }
func (frame *FrameHTTPResBody) Kind () FrameKind { return FrameKindHTTPResBody }
func (frame *FrameHTTPResEnd)  Kind () FrameKind { return FrameKindHTTPResEnd  }

// frame data getters
func (frame *FrameHTTPReqBody) GetData () []byte { return frame.Data }
func (frame *FrameHTTPResBody) GetData () []byte { return frame.Data }

/* ParseFrame splits a frame into its kind and its data. Unmarshaling should be
 * conducted by the handler.
 */
func ParseFrame (frameData []byte) (kind FrameKind, data []byte, err error) {
        if len(frameData) < 1 { return 0, nil, errors.New("empty frame") }
        return FrameKind(frameData[0]), frameData[1:], nil
}

/* ReadParseFrame reads a frame from an fsock reader and parses it.
 */
func ReadParseFrame (
        reader *fsock.Reader,
) (
        kind FrameKind,
        data []byte,
        err error,
) {
        frame, err := reader.Read()
        if err != nil { return 0, nil, err }
        return ParseFrame(frame)
}

/* MarshalFrame takes in a struct satisfying the Frame interface and endcodes it
 * into a valid frame.
 */
func MarshalFrame (
        frame Frame,
) (
        frameData []byte,
        err       error,
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
                if err != nil { return nil, err }
                break
        }
        
        frameData = append(data, 0)
        copy(frameData[1:], frameData)
        frameData[0] = byte(frame.Kind())
        return
}

/* WriteMarshalFrame marshals and writes a Frame.
 */
func WriteMarshalFrame (writer *fsock.Writer, frame Frame) (nn int, err error) {
        frameData, err := MarshalFrame(frame)
        if err != nil { return 0, err }
        return writer.WriteFrame(frameData)
}
