package codec

type Profile struct {
	Name        string
	Resolution  string
	MaxBitrate  int // kbps
	FrameRate   int
	CodecParams string
}

type Ladder struct {
	profiles []Profile
}

func NewLadder() *Ladder {
	return &Ladder{
		profiles: []Profile{
			{Name: "H.264", Resolution: "720p", MaxBitrate: 5000, FrameRate: 60, CodecParams: "preset=ultrafast"},
			{Name: "H.264", Resolution: "1080p", MaxBitrate: 15000, FrameRate: 60, CodecParams: "preset=veryfast"},
			{Name: "AV1", Resolution: "1080p", MaxBitrate: 18000, FrameRate: 60, CodecParams: "cpu-used=4"},
			{Name: "HEVC", Resolution: "1080p", MaxBitrate: 20000, FrameRate: 60, CodecParams: "preset=fast"},
			{Name: "HEVC", Resolution: "1440p", MaxBitrate: 35000, FrameRate: 60, CodecParams: "preset=medium"},
			{Name: "AV1", Resolution: "4K", MaxBitrate: 40000, FrameRate: 60, CodecParams: "cpu-used=4"},
			{Name: "HEVC", Resolution: "4K", MaxBitrate: 50000, FrameRate: 60, CodecParams: "preset=medium"},
		},
	}
}

func (l *Ladder) GetProfiles() []Profile {
	return l.profiles
}

func (l *Ladder) GetProfileForBandwidth(bandwidthKbps int) Profile {
	// Walk down the ladder to find the best profile for given bandwidth
	for i := len(l.profiles) - 1; i >= 0; i-- {
		if bandwidthKbps >= l.profiles[i].MaxBitrate {
			return l.profiles[i]
		}
	}
	// Return lowest profile if bandwidth is too low
	return l.profiles[0]
}
