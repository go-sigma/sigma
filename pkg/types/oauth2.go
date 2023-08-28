// Copyright 2023 sigma
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import "github.com/go-sigma/sigma/pkg/types/enums"

// Oauth2LoginRequest represents the request to login.
type Oauth2LoginRequest struct {
	Provider enums.Provider `json:"provider" param:"provider" validate:"required,is_valid_provider"`
}

// Oauth2CallbackRequest represents the request to callback.
type Oauth2CallbackRequest struct {
	Provider enums.Provider `json:"provider" param:"provider" validate:"required,is_valid_provider" example:"gitlab"`
	Code     string         `json:"code" query:"code" validate:"required" example:"123456"`
	Endpoint string         `json:"endpoint" query:"endpoint" example:"http://localhost:5173"`
}

// Oauth2RedirectCallbackRequest ...
type Oauth2RedirectCallbackRequest struct {
	Code string `json:"code" query:"code"`
}

// Oauth2UserInfo represents the user info.
type Oauth2UserInfo struct {
	Provider     enums.Provider `json:"provider"`
	ID           string         `json:"id"`
	Username     string         `json:"username"`
	Email        string         `json:"email"`
	Token        string         `json:"token"`
	RefreshToken string         `json:"refresh_token"`
}

// Oauth2CallbackResponse ...
type Oauth2CallbackResponse struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJYSW1hZ2VyIiwic3ViIjoieGltYWdlciIsImV4cCI6MTY4OTUwNDcxMSwibmJmIjoxNjg5NTAxMTExLCJpYXQiOjE2ODk1MDExMTEsImp0aSI6ImIxYTRmMGQ1LWNlNTUtNGU3ZS1hYjY1LWRhNTZiMDhhNjg2YSIsImRhdCI6InhpbWFnZXIifQ.cjuIJVnLl9yFB9zhv0KPZUq_M1Mb-tiyjHQXYowRAROEdu5t6HHYnFnl348IYFg51vGDh7ROBp1-pZIQJ5gCyM4rTuoYZneS7NPtb8sFjch3dDotVDXSpbppdkXZAPvEIwXKDKcmyMCsAAgep4A6gVeQ07RthUbITahCG3-ssF9NTojDgIyKysReju3BV5FOh_lbBwNXmfBnRUV11w8eApAuLEJhdNM_W50BdoHvUHAbwblDmanNonc9zAkzcQQZqndCNZJJ2hee7ZqOSByWDtxnLB5zbvLBV76BJf6EAW8zTYDW9fxWwSydhvmo5bSxgcI4LFzloUXO-Mj1TWVg2Lvjf3vAkmdYUxD8fhxE7x49i02TN4ohwtr3jI27vOh4Jv4FgMbu2SkZTVrfQ7ySpcWgX-UC2egXSs2fwpwoPyDZn4LmnDTZX4_PLqz7IgoeusrpFzHnfKD_mf3q-xq1ugJoNQRFWXFpF9fhWmYPsefoKlU349ZVqHg19QT2sFnSJBHWqL92NAr75vzUxxmxN61ZpXU70xZ54-qXMsu1V2jyGQl2wlFDPPb8jUWEh9cY_EmEarFAJPCBTAaxhdTpe8lR7b4WcbHtGu2zDQYpDvNOL7NKTLzjzn1COewvE2jkf0m9fL-u3RzrEIfo4eLBSBbUrrnpFit7CQFzxUZF5u9IWaSuicqwy7KoFt3PazsvQNYi9DYoGi5TVuI2EtdWYCSA09J5rL3GKkUkwZT0yMrea21xR9tpBU4LvJLM00bXYXLQGwISoSQ30pLGJiOskDADMrF-Wfg7JZi1KiUyA8jNgNebOw9VVBYxR7h33vKNDJPI2dZsqOYAwXqaTQTdJAm888yrpBRt22s2lsWhUBmvRgHpDFUHKUQHTFNZNOi_CeL4YTaoWhcS9j6ydtrteDz3gw783hY9_kSnER0GiYZNyMPMJYcQTteeESwCP0_eRgZDtc7jFU2ZDFSWshzWk7M53YQvuSw9j3r5l3yJ88qYLgJoqnLgGBHOfdz5zkzJkECEXzbmb05JB7cnJUNgg_AJSpI38P7906JBXsBmgXpjqDyFdYn89NbqGZqwcyKEquvEtDfdSAIAqlbVT-g8lkC14T3YD-CJwhK7u3lB-bFAASOdb4xjz5hcL9C7KoElMNGxuK0r-7bDYBVqQVSt_jqbAPufx3fgpz8D-S-43DkN7ZIZTCaLrocNZjgT74KXlCzBYnPgTAOvPxOPzOUxrwgXLKbloKSWSAr8eOEgR2bdF8WFI7NG6WunlJp55v9yc5KTTeuaoDQGZhuVzAH5A05NtDDerT7KHeoiI2_q9s_VrY6J1er6bWq2VI46iYl339ozgje10RCDWCbWzWiGeg"`
	Token        string `json:"token" example:"eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJYSW1hZ2VyIiwic3ViIjoieGltYWdlciIsImV4cCI6MTY4OTUwNDcxMSwibmJmIjoxNjg5NTAxMTExLCJpYXQiOjE2ODk1MDExMTEsImp0aSI6ImIxYTRmMGQ1LWNlNTUtNGU3ZS1hYjY1LWRhNTZiMDhhNjg2YSIsImRhdCI6InhpbWFnZXIifQ.cjuIJVnLl9yFB9zhv0KPZUq_M1Mb-tiyjHQXYowRAROEdu5t6HHYnFnl348IYFg51vGDh7ROBp1-pZIQJ5gCyM4rTuoYZneS7NPtb8sFjch3dDotVDXSpbppdkXZAPvEIwXKDKcmyMCsAAgep4A6gVeQ07RthUbITahCG3-ssF9NTojDgIyKysReju3BV5FOh_lbBwNXmfBnRUV11w8eApAuLEJhdNM_W50BdoHvUHAbwblDmanNonc9zAkzcQQZqndCNZJJ2hee7ZqOSByWDtxnLB5zbvLBV76BJf6EAW8zTYDW9fxWwSydhvmo5bSxgcI4LFzloUXO-Mj1TWVg2Lvjf3vAkmdYUxD8fhxE7x49i02TN4ohwtr3jI27vOh4Jv4FgMbu2SkZTVrfQ7ySpcWgX-UC2egXSs2fwpwoPyDZn4LmnDTZX4_PLqz7IgoeusrpFzHnfKD_mf3q-xq1ugJoNQRFWXFpF9fhWmYPsefoKlU349ZVqHg19QT2sFnSJBHWqL92NAr75vzUxxmxN61ZpXU70xZ54-qXMsu1V2jyGQl2wlFDPPb8jUWEh9cY_EmEarFAJPCBTAaxhdTpe8lR7b4WcbHtGu2zDQYpDvNOL7NKTLzjzn1COewvE2jkf0m9fL-u3RzrEIfo4eLBSBbUrrnpFit7CQFzxUZF5u9IWaSuicqwy7KoFt3PazsvQNYi9DYoGi5TVuI2EtdWYCSA09J5rL3GKkUkwZT0yMrea21xR9tpBU4LvJLM00bXYXLQGwISoSQ30pLGJiOskDADMrF-Wfg7JZi1KiUyA8jNgNebOw9VVBYxR7h33vKNDJPI2dZsqOYAwXqaTQTdJAm888yrpBRt22s2lsWhUBmvRgHpDFUHKUQHTFNZNOi_CeL4YTaoWhcS9j6ydtrteDz3gw783hY9_kSnER0GiYZNyMPMJYcQTteeESwCP0_eRgZDtc7jFU2ZDFSWshzWk7M53YQvuSw9j3r5l3yJ88qYLgJoqnLgGBHOfdz5zkzJkECEXzbmb05JB7cnJUNgg_AJSpI38P7906JBXsBmgXpjqDyFdYn89NbqGZqwcyKEquvEtDfdSAIAqlbVT-g8lkC14T3YD-CJwhK7u3lB-bFAASOdb4xjz5hcL9C7KoElMNGxuK0r-7bDYBVqQVSt_jqbAPufx3fgpz8D-S-43DkN7ZIZTCaLrocNZjgT74KXlCzBYnPgTAOvPxOPzOUxrwgXLKbloKSWSAr8eOEgR2bdF8WFI7NG6WunlJp55v9yc5KTTeuaoDQGZhuVzAH5A05NtDDerT7KHeoiI2_q9s_VrY6J1er6bWq2VI46iYl339ozgje10RCDWCbWzWiGeg"`
	ID           int64  `json:"id" example:"1"`
	Email        string `json:"email" example:"test@email.com"`
	Username     string `json:"username" example:"sigma"`
}

// Oauth2ClientIDRequest ...
type Oauth2ClientIDRequest struct {
	Provider enums.Provider `json:"provider" param:"provider" validate:"required,is_valid_provider"`
}

// Oauth2ClientIDResponse ...
type Oauth2ClientIDResponse struct {
	ClientID string `json:"client_id" example:"1234567890"`
}
