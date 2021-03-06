/*
 * Copyright 2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

type Windowing struct {
	// When set to some value N, uses windows of N input messages.
	Size int32 `json:"size,omitempty"`

	// When set to some value D, will slice windows every D (uses time.Duration#Parse()).
	Time string `json:"time,omitempty"`

	// When set to some value D, slices windows with inactivity period of D (uses time.Duration#Parse()).
	Session string `json:"session,omitempty"`
}

func (w *Windowing) IsUnbounded() bool {
	return w.Size == 0 && w.Time == "" && w.Session == ""
}
