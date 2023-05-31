/*
 * Copyright 1999-2020 Alibaba Group Holding Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package azure

import (
	"github.com/chaosblade-io/chaosblade-spec-go/spec"
)

type AzureCommandSpec struct {
	spec.BaseExpModelCommandSpec
}

func NewAzureCommandSpec() spec.ExpModelCommandSpec {
	return &AzureCommandSpec{
		spec.BaseExpModelCommandSpec{
			ExpActions: []spec.ExpActionCommandSpec{
				NewVmActionSpec(),
			},
			ExpFlags: []spec.ExpFlagSpec{},
		},
	}
}

func (*AzureCommandSpec) Name() string {
	return "azure"
}

func (*AzureCommandSpec) ShortDesc() string {
	return "Azure experiment"
}

func (*AzureCommandSpec) LongDesc() string {
	return "Azure experiment contains vm start, stop, disk loss"
}
