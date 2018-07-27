// Copyright 2018 The eballscan Authors
// This file is part of the eballscan.
//
// The eballscan is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The eballscan is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the eballscan. If not, see <http://www.gnu.org/licenses/>.

package database

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/ecoball/eballscan/data"
	"github.com/ecoball/eballscan/syn"
	"github.com/ecoball/go-ecoball/common/elog"
	_ "github.com/lib/pq"
)

var (
	CockroachDb *sql.DB
	DbMutex     sync.Mutex
	log         = elog.NewLogger("database", elog.DebugLog)
)

func init() {
	// Connect to the "bank" database.
	var err error
	CockroachDb, err = sql.Open("postgres", "postgresql://eballscan@localhost:26257/blockchain?sslmode=disable")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}

	// Create the "blocks" table.
	if _, err = CockroachDb.Exec(
		`create table if not exists blocks (hight int primary key, 
			hash varchar(70), prevHash varchar(70), merkleHash varchar(70), stateHash varchar(70), countTxs int)`); err != nil {
		log.Fatal(err)
	}

	// Print out the balances.
	rows, errQuery := CockroachDb.Query("select hight, hash, prevHash, merkleHash, stateHash, countTxs from blocks")
	if errQuery != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			hight, countTxs                       int
			hash, prevHash, merkleHash, stateHash string
		)

		if err := rows.Scan(&hight, &hash, &prevHash, &merkleHash, &stateHash, &countTxs); err != nil {
			log.Fatal(err)
		}

		data.Blocks.Add(hight, data.BlockInfo{hash, prevHash, merkleHash, stateHash, countTxs})

		if hight > syn.MaxHight {
			syn.MaxHight = hight
		}
	}
}

func AddBlock(hight, countTxs int, hash, prevHash, merkleHash, stateHash string) error {
	DbMutex.Lock()
	defer DbMutex.Unlock()

	var values string
	values = fmt.Sprintf(`(%d, '%s', '%s', '%s', '%s', %d)`, hight, hash, prevHash, merkleHash, stateHash, countTxs)
	values = "insert into blocks(hight, hash, prevHash, merkleHash, stateHash, countTxs) values" + values
	_, err := CockroachDb.Exec(values)
	if nil != err {
		return err
	}

	data.Blocks.Add(hight, data.BlockInfo{hash, prevHash, merkleHash, stateHash, countTxs})
	return nil
}
