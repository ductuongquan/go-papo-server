// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package facebookgraph

type FacebookImageData struct {
	Height 							int    					`json:"height"`
	Width  							int    					`json:"width"`
	Src 							string					`json:"src,omitempty"`
	Url								string 					`json:"url,omitempty"` // không có ở comment attachment
	PreviewUrl 						string 					`json:"preview_url,omitempty"` // không có ở comment attachment
	ImageType 						int 					`json:"image_type,omitempty"` // không có ở comment attachment
	RenderAsSticker					bool 					`json:"render_as_sticker,omitempty"` // không có ở comment attachment
	MaxWidth						string 					`json:"max_width,omitempty"`
	MaxHeight 						string 					`json:"max_height,omitempty"`
	RawGifImage						string 					`json:"raw_gif_image,omitempty"`
	AnimatedGifUrl					string 					`json:"animated_gif_url,omitempty"`
	AnimatedGifPreviewUrl			string 					`json:"animated_gif_preview_url,omitempty"`
}