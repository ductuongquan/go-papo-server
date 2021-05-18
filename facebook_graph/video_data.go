package facebookgraph

type FacebookVideoData struct {
	Height 							int    					`json:"height"`
	Width  							int    					`json:"width"`
	Url								string 					`json:"url,omitempty"` // không có ở comment attachment
	PreviewUrl 						string 					`json:"preview_url,omitempty"` // không có ở comment attachment
	Length 							int 					`json:"length"`
	VideoType 						int 					`json:"video_type"`
	Rotation 						int 					`json:"rotation"`
}
