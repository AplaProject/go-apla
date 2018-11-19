package metadata

type teststruct struct {
	Key    int
	Value1 string
	Value2 []byte
}

//func TestRollbackSaveState(t *testing.T) {
//	txMock := &kv.MockTransaction{}
//	mr := rollback{tx: txMock, counter: counter{txCounter: make(map[string]uint64)}}
//
//	registry := &types.Registry{
//		Name:      "keys",
//		Ecosystem: &types.Ecosystem{Name: "aaa"},
//	}
//
//	block, tx := []byte("123"), []byte("321")
//
//	s := state{Transaction: string(tx), Counter: 1, RegistryName: registry.Name, Ecosystem: registry.Ecosystem.Name, Key: "1"}
//	jstate, err := json.Marshal(s)
//	require.Nil(t, err)
//	txMock.On("Set", fmt.Sprintf(writePrefix, string(block), 1, string(tx)), string(jstate)).Return(nil)
//	require.Nil(t, mr.saveState(block, tx, registry, "1", ""))
//	require.Equal(t, mr.counter.txCounter[string(block)], uint64(1))
//
//	structValue := teststruct{
//		Key:    666,
//		Value1: "stringvalue",
//		Value2: make([]byte, 20),
//	}
//	jsonValue, err := json.Marshal(structValue)
//	require.Nil(t, err)
//	s = state{Transaction: string(tx), Counter: 2, RegistryName: registry.Name, Ecosystem: registry.Ecosystem.Name, Value: string(jsonValue), Key: "2"}
//	jstate, err = json.Marshal(s)
//	require.Nil(t, err)
//	txMock.On("Set", fmt.Sprintf(writePrefix, string(block), 2, string(tx)), string(jstate)).Return(nil)
//	require.Nil(t, mr.saveState(block, tx, registry, "2", string(jsonValue)))
//	require.Equal(t, mr.counter.txCounter[string(block)], uint64(2))
//
//	s = state{Transaction: string(tx), Counter: 3, RegistryName: registry.Name, Ecosystem: registry.Ecosystem.Name, Value: "", Key: "3"}
//	jstate, err = json.Marshal(s)
//	require.Nil(t, err)
//	txMock.On("Set", fmt.Sprintf(writePrefix, string(block), 3, string(tx)), string(jstate)).Return(errors.New("testerr"))
//	require.Error(t, mr.saveState(block, tx, registry, "3", ""))
//	require.Equal(t, mr.counter.txCounter[string(block)], uint64(2))
//}
