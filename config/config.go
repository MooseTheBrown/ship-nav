package config

import (
	"encoding/json"
	"io"
	"os"
)

type coreConfig struct {
	Declination          float64 `json:"declination"`
	UpdateBufSize        int     `json:"updateBufSize"`
	TurningSpeed         string  `json:"turningSpeed"`
	TurningSteeringLeft  string  `json:"turningSteeringLeft"`
	TurningSteeringRight string  `json:"turningSteeringRight"`
	ApproachSpeed        string  `json:"approachSpeed"`
	FullSpeed            string  `json:"fullSpeed"`
	ApproachDistance     float64 `json:"approachDistance"`
	DistanceInaccuracy   float64 `json:"distanceInaccuracy"`
}

type networkConfig struct {
	SocketName string `json:"socketName"`
}

type positionConfig struct {
	SocketName      string `json:"socketName"`
	PollingInterval int64  `json:"pollingInterval"`
}

type shipConfig struct {
	SocketName      string `json:"socketName"`
	PollingInterval int64  `json:"pollingInterval"`
}

type Config struct {
	CoreConfig     *coreConfig     `json:"coreConfig"`
	NetworkConfig  *networkConfig  `json:"networkConfig"`
	PositionConfig *positionConfig `json:"positionConfig"`
	ShipConfig     *shipConfig     `json:"shipConfig"`
	LogLevel       string          `json:"logLevel"`
}

func NewConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	configData, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Declination() float64 {
	return c.CoreConfig.Declination
}

func (c *Config) UpdateBufSize() int {
	return c.CoreConfig.UpdateBufSize
}

func (c *Config) TurningSpeed() string {
	return c.CoreConfig.TurningSpeed
}

func (c *Config) TurningSteeringLeft() string {
	return c.CoreConfig.TurningSteeringLeft
}

func (c *Config) TurningSteeringRight() string {
	return c.CoreConfig.TurningSteeringRight
}

func (c *Config) ApproachSpeed() string {
	return c.CoreConfig.ApproachSpeed
}

func (c *Config) FullSpeed() string {
	return c.CoreConfig.FullSpeed
}

func (c *Config) ApproachDistance() float64 {
	return c.CoreConfig.ApproachDistance
}

func (c *Config) DistanceInaccuracy() float64 {
	return c.CoreConfig.DistanceInaccuracy
}

func (c *Config) NetworkSocketName() string {
	return c.NetworkConfig.SocketName
}

func (c *Config) PositionSocketName() string {
	return c.PositionConfig.SocketName
}

func (c *Config) PositionPollingInterval() int64 {
	return c.PositionConfig.PollingInterval
}

func (c *Config) ShipSocketName() string {
	return c.ShipConfig.SocketName
}

func (c *Config) ShipPollingInterval() int64 {
	return c.ShipConfig.PollingInterval
}
