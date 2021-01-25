package contract_test

func removeKey(key string, defaultData map[string]string) map[string]string {
	temp := make(map[string]string)

	for k, v := range defaultData {
		if k != key {
			temp[k] = v
		}
	}

	return temp
}
