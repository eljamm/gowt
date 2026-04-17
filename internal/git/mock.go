package git

type MockCommander struct {
	WorktreeListOutput string
	WorktreeListErr    error
	RevParseOutput     string
	RevParseErr        error
	GitCommonDirOutput string
	GitCommonDirErr    error
	WorktreeRemoveErr  error
	WorktreeAddErr     error
	WorktreeAddNewErr  error
	BranchDeleteErr    error
}

func (m *MockCommander) WorktreeList() (string, error) {
	return m.WorktreeListOutput, m.WorktreeListErr
}

func (m *MockCommander) RevParse(showToplevel bool) (string, error) {
	return m.RevParseOutput, m.RevParseErr
}

func (m *MockCommander) GitCommonDir() (string, error) {
	return m.GitCommonDirOutput, m.GitCommonDirErr
}

func (m *MockCommander) WorktreeRemove(path string, force bool) error {
	return m.WorktreeRemoveErr
}

func (m *MockCommander) WorktreeAdd(path, branch string) error {
	return m.WorktreeAddErr
}

func (m *MockCommander) WorktreeAddNew(path, branch string) error {
	return m.WorktreeAddNewErr
}

func (m *MockCommander) BranchDelete(branch string, force bool) error {
	return m.BranchDeleteErr
}
