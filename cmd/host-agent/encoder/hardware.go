package encoder

type HardwareEncoder interface {
    Name() string
    Encode(frame []byte) ([]byte, error)
    Close() error
}

type nvencEncoder struct {
    name string
}

func (n *nvencEncoder) Name() string {
    return n.name
}

func (n *nvencEncoder) Encode(frame []byte) ([]byte, error) {
    // Stub: NVENC encoding
    return append([]byte{0x01}, frame...), nil
}

func (n *nvencEncoder) Close() error {
    return nil
}

type qsvEncoder struct {
    name string
}

func (q *qsvEncoder) Name() string {
    return q.name
}

func (q *qsvEncoder) Encode(frame []byte) ([]byte, error) {
    // Stub: Intel Quick Sync Video
    return append([]byte{0x02}, frame...), nil
}

func (q *qsvEncoder) Close() error {
    return nil
}

type amfEncoder struct {
    name string
}

func (a *amfEncoder) Name() string {
    return a.name
}

func (a *amfEncoder) Encode(frame []byte) ([]byte, error) {
    // Stub: AMD AMF
    return append([]byte{0x03}, frame...), nil
}

func (a *amfEncoder) Close() error {
    return nil
}

type videoToolboxEncoder struct {
    name string
}

func (v *videoToolboxEncoder) Name() string {
    return v.name
}

func (v *videoToolboxEncoder) Encode(frame []byte) ([]byte, error) {
    // Stub: VideoToolbox (macOS)
    return append([]byte{0x04}, frame...), nil
}

func (v *videoToolboxEncoder) Close() error {
    return nil
}

type vaapiEncoder struct {
    name string
}

func (v *vaapiEncoder) Name() string {
    return v.name
}

func (v *vaapiEncoder) Encode(frame []byte) ([]byte, error) {
    // Stub: VAAPI (Linux)
    return append([]byte{0x05}, frame...), nil
}

func (v *vaapiEncoder) Close() error {
    return nil
}

func NewHardwareEncoder(encoderType string) HardwareEncoder {
    switch encoderType {
    case "nvenc":
        return &nvencEncoder{name: "NVENC"}
    case "qsv":
        return &qsvEncoder{name: "Intel QSV"}
    case "amf":
        return &amfEncoder{name: "AMD AMF"}
    case "videotoolbox":
        return &videoToolboxEncoder{name: "VideoToolbox"}
    case "vaapi":
        return &vaapiEncoder{name: "VAAPI"}
    default:
        return nil
    }
}
