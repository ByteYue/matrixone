// Copyright 2022 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package engine

import (
	"context"

	"github.com/matrixorigin/matrixone/pkg/vm/engine"
)

type Database struct {
}

var _ engine.Database = new(Database)

func (*Database) Create(ctx context.Context, relName string, defs []engine.TableDef) error {
	//TODO
	return nil
}

func (*Database) Delete(ctx context.Context, relName string) error {
	//TODO
	return nil
}

func (*Database) Relation(ctx context.Context, relName string) (engine.Relation, error) {
	//TODO
	return nil, nil
}

func (*Database) Relations(ctx context.Context) ([]string, error) {
	//TODO
	return nil, nil
}
