package models

type Torrent struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Size          int64    `json:"size"`
	Progress      float64  `json:"progress"`
	Status        string   `json:"status"`
	DownloadSpeed int64    `json:"download_speed"`
	UploadSpeed   int64    `json:"upload_speed"`
	MagnetLink    string   `json:"magnet_link"`
	InfoHash      [20]byte `json:"info_hash"`
}
