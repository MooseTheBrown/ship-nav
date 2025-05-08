package position

const (
	CmdGetGPS           = "GetGPSData"
	CmdGetMagnetometer  = "GetMagnetometerData"
	CmdStartCalibration = "StartCalibration"
	CmdStopCalibration  = "StopCalibration"
)

type IPCRequest struct {
	Cmd string `json:"cmd"`
}

type GPSInfoResponse struct {
	NumSatellites int     `json:"numSatellites"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	SpeedKnots    float64 `json:"speedKnots"`
	SpeedKm       float64 `json:"speedKm"`
}

type MagnetometerInfoResponse struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
	Z int32 `json:"z"`
}

type CalibrationResponse struct {
	Success bool `json:"success"`
}

type ErrorResponse struct {
	ErrorMessage string `json:"error_message"`
}
