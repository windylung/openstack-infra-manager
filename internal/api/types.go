package api

type ApplyQuotaRequest struct {
	ProjectID string `json:"projectId"`

	Nova *struct {
		Cores     *int `json:"cores"`
		RAMMB     *int `json:"ramMB"`
		Instances *int `json:"instances"`
	} `json:"nova,omitempty"`

	Cinder *struct {
		Volumes   *int `json:"volumes"`
		Snapshots *int `json:"snapshots"`
		Gigabytes *int `json:"gigabytes"`
	} `json:"cinder,omitempty"`
}
