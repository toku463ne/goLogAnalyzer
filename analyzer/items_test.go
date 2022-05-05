package analyzer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/pkg/errors"
)

func Test_Items(t *testing.T) {
	blockNoIdx := getColIdx("circuitDBStatus", "blockNo")
	itemIdx := getColIdx("items", "item")

	registerTran := func(it *items, itemCount int, a ...string) error {
		for _, item := range a {
			if itemID := it.register(item, itemCount, 0, true); itemID < 0 {
				return errors.New("Failed to register the item " + item)
			}
		}
		if err := it.next(); err != nil {
			return err
		}
		return nil
	}
	registerTrans := func(it *items, a [][]string) error {
		for _, tran := range a {
			if err := registerTran(it, 1, tran...); err != nil {
				return err
			}
		}
		return nil
	}

	blockExists := func(it *items, blockNo int) bool {
		cnt := it.statusTable.Count(func(v []string) bool {
			return v[blockNoIdx] == strconv.Itoa(blockNo)
		})
		return cnt > 0
	}

	checkCircuitDBStatus := func(it *items, blockNo int, expected []interface{}) error {
		if !blockExists(it, blockNo) {
			return errors.New("blockNo must be uniq")
		}

		var lastIndex int
		var blockID string
		var rowNo int
		var completed bool
		if err := it.statusTable.Select1Row(func(v []string) bool {
			return v[blockNoIdx] == strconv.Itoa(blockNo)
		}, []string{"lastIndex", "blockID", "rowNo", "completed"},
			&lastIndex, &blockID, &rowNo, &completed); err != nil {
			return err
		}

		if err := getGotExpErr("lastIndex", lastIndex, expected[0]); err != nil {
			return err
		}
		if err := getGotExpErr("blockID", blockID, expected[1]); err != nil {
			return err
		}
		if err := getGotExpErr("rowNo", rowNo, expected[2]); err != nil {
			return err
		}
		if err := getGotExpErr("completed", completed, expected[3]); err != nil {
			return err
		}

		return nil
	}

	getItemCountInBlock := func(it *items, blockNo int, item string) int {
		tableName := it.getBlockTableName(blockNo)
		if !blockExists(it, blockNo) {
			return 0
		}
		g, err := it.GetGroup("items")
		if err != nil {
			return -1
		}
		t, err := g.GetTable(tableName)
		if err != nil {
			return -1
		}
		itemCount := 0
		if err := t.Sum(func(v []string) bool {
			return v[itemIdx] == item
		}, "itemCount", &itemCount); err != nil {
			return -1
		}

		return itemCount
	}

	checkItemCountInBlock := func(it *items, blockNo int, item string, expCnt int) error {
		cnt := getItemCountInBlock(it, blockNo, item)
		if cnt != expCnt {
			if err := getGotExpErr(item, cnt, expCnt); err != nil {
				return err
			}
		}
		return nil
	}

	checkItemCount := func(it *items, item string, expCnt int) error {
		itemID := it.getItemID(item)
		if itemID == -1 {
			return errors.New(fmt.Sprintf("item %s is not registered.", item))
		}
		cnt := it.getCount(itemID)
		if cnt != expCnt {
			if err := getGotExpErr(item, cnt, expCnt); err != nil {
				return err
			}
		}
		return nil
	}

	dataDir, err := initTestDir("itemsTest")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	maxBlocks := 3
	maxRowsInBlock := 5
	useGzipInCircuitTables = false
	//tl, err := newTableLogRecords(dataDir, maxBlocks, maxRowsInBlock)
	it, err := newItems(dataDir, "items", maxBlocks, maxRowsInBlock)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	it.DropAll()
	it, err = newItems(dataDir, "items", maxBlocks, maxRowsInBlock)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	inRows := [][]string{
		{"test100", "test200", "test301"},
		{"test100", "test200", "test302"},
		{"test100", "test200", "test303"},
	}

	if err := registerTrans(it, inRows); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkItemCount(it, "test100", 3); err != nil {
		t.Errorf("%v", err)
		return
	}

	inRows = [][]string{
		{"test100", "test200", "test304"},
		{"test100", "test200", "test305"},
		{"test100", "test201", "test306"},
	}

	if err := registerTrans(it, inRows); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkItemCount(it, "test100", 6); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkItemCount(it, "test200", 5); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkItemCountInBlock(it, 0, "test100", 5); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkItemCountInBlock(it, 0, "test200", 5); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkCircuitDBStatus(it, 0,
		[]interface{}{0, "BLK0000000000", 5, true}); err != nil {
		t.Errorf("%v", err)
		return
	}

	inRows = [][]string{
		{"test100", "test201", "test307"},
		{"test100", "test201", "test308"},
		{"test100", "test201", "test309"},
		{"test100", "test201", "test310"},
		{"test100", "test202", "test311"},
		{"test100", "test202", "test312"},
		{"test100", "test202", "test313"},
		{"test100", "test202", "test314"},
		{"test100", "test202", "test315"},
		{"test100", "test203", "test316"},
	}

	if err := registerTrans(it, inRows); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkItemCount(it, "test100", 16); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkItemCount(it, "test200", 5); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkCircuitDBStatus(it, 0,
		[]interface{}{0, "BLK0000000000", 5, true}); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkCircuitDBStatus(it, 1,
		[]interface{}{1, "BLK0000000001", 5, true}); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkCircuitDBStatus(it, 2,
		[]interface{}{2, "BLK0000000002", 5, true}); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkItemCountInBlock(it, 0, "test100", 5); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkItemCountInBlock(it, 0, "test200", 5); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkItemCountInBlock(it, 1, "test100", 5); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkItemCountInBlock(it, 1, "test200", 0); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkItemCountInBlock(it, 1, "test201", 5); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkItemCountInBlock(it, 2, "test100", 5); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkItemCountInBlock(it, 2, "test202", 5); err != nil {
		t.Errorf("%v", err)
		return
	}

	err = it.commit(false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkItemCount(it, "test100", 16); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkItemCount(it, "test200", 5); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkItemCount(it, "test301", 1); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkCircuitDBStatus(it, 0,
		[]interface{}{3, "BLK0000000000", 1, false}); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkItemCountInBlock(it, 0, "test100", 1); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkItemCountInBlock(it, 0, "test200", 0); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkItemCountInBlock(it, 0, "test203", 1); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkItemCountInBlock(it, 0, "test316", 1); err != nil {
		t.Errorf("%v", err)
		return
	}

	it = nil

	it, err = newItems(dataDir, "items", maxBlocks, maxRowsInBlock)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := it.load(); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkItemCount(it, "test100", 11); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkItemCount(it, "test203", 1); err != nil {
		t.Errorf("%v", err)
		return
	}

	inRows = [][]string{
		{"test100", "test203", "test317"},
		{"test100", "test203", "test318"},
		{"test100", "test203", "test319"},
		{"test100", "test203", "test320"},
		{"test100", "test204", "test321"},
	}

	if err := registerTrans(it, inRows); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkItemCount(it, "test100", 16); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkItemCount(it, "test203", 5); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkCircuitDBStatus(it, 0,
		[]interface{}{3, "BLK0000000000", 5, true}); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkItemCountInBlock(it, 0, "test100", 5); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkItemCountInBlock(it, 0, "test203", 5); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkItemCountInBlock(it, 0, "test204", 0); err != nil {
		t.Errorf("%v", err)
		return
	}

	it = nil
}
