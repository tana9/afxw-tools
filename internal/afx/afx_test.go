package afx

import (
	"errors"
	"reflect"
	"testing"
	"unsafe"

	"github.com/go-ole/go-ole"
)

func stubCallMethod(t *testing.T, stub func(*ole.IDispatch, string, ...interface{}) (*ole.VARIANT, error)) {
	t.Helper()
	original := callCOMMethod
	callCOMMethod = stub
	t.Cleanup(func() { callCOMMethod = original })
}

func stringVariant(value string) ole.VARIANT {
	bstr := ole.SysAllocString(value)
	return ole.NewVariant(ole.VT_BSTR, int64(uintptr(unsafe.Pointer(bstr))))
}

func TestClose_NilFields(t *testing.T) {
	// afxw, unknown が両方 nil の状態で Close() を呼んでもパニックしないことを確認
	a := &oleClient{afxw: nil, unknown: nil}
	a.Close()
}

func TestEnsureTrailingBackslash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`C:\Users\Test`, `C:\Users\Test\`},
		{`C:\Users\Test\`, `C:\Users\Test\`},
		{`C:\`, `C:\`},
		{`C:`, `C:\`},
		{``, `\`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ensureTrailingBackslash(tt.input)
			if got != tt.expected {
				t.Errorf("期待: %q, 取得: %q", tt.expected, got)
			}
		})
	}
}

func TestChangeDirectory(t *testing.T) {
	t.Run("calls Exec and clears result", func(t *testing.T) {
		result := ole.NewVariant(ole.VT_I4, 1)
		stubCallMethod(t, func(_ *ole.IDispatch, name string, params ...interface{}) (*ole.VARIANT, error) {
			if name != "Exec" {
				t.Fatalf("method = %q, want Exec", name)
			}
			want := `&EXCD -P"C:\Work\"`
			if len(params) != 1 || params[0] != want {
				t.Fatalf("params = %#v, want [%q]", params, want)
			}
			return &result, nil
		})

		if err := (&oleClient{}).ChangeDirectory(`C:\Work`); err != nil {
			t.Fatalf("ChangeDirectory() error = %v", err)
		}
		if result.VT != ole.VT_EMPTY {
			t.Errorf("result was not cleared: VT = %v", result.VT)
		}
	})

	t.Run("wraps call error", func(t *testing.T) {
		callErr := errors.New("exec failed")
		stubCallMethod(t, func(_ *ole.IDispatch, _ string, _ ...interface{}) (*ole.VARIANT, error) {
			return nil, callErr
		})

		err := (&oleClient{}).ChangeDirectory(`C:\Work`)
		if !errors.Is(err, callErr) {
			t.Fatalf("ChangeDirectory() error = %v, want wrapped %v", err, callErr)
		}
	})
}

func TestDirectoryHistories(t *testing.T) {
	var results []*ole.VARIANT
	stubCallMethod(t, func(_ *ole.IDispatch, name string, params ...interface{}) (*ole.VARIANT, error) {
		var result ole.VARIANT
		switch name {
		case "HisDirCount":
			if !reflect.DeepEqual(params, []interface{}{WindowLeft}) {
				t.Fatalf("HisDirCount params = %#v", params)
			}
			result = ole.NewVariant(ole.VT_I4, 2)
		case "HisDir":
			index := params[1].(int)
			result = stringVariant([]string{`C:\First`, `C:\Second`}[index])
		default:
			t.Fatalf("unexpected method %q", name)
		}
		results = append(results, &result)
		return &result, nil
	})

	got, err := (&oleClient{}).DirectoryHistories([]int{WindowLeft})
	if err != nil {
		t.Fatalf("DirectoryHistories() error = %v", err)
	}
	want := []string{`C:\First`, `C:\Second`}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DirectoryHistories() = %v, want %v", got, want)
	}
	for i, result := range results {
		if result.VT != ole.VT_EMPTY {
			t.Errorf("result[%d] was not cleared: VT = %v", i, result.VT)
		}
	}
}

func TestDirectoryHistoriesErrors(t *testing.T) {
	t.Run("count call", func(t *testing.T) {
		callErr := errors.New("count failed")
		stubCallMethod(t, func(_ *ole.IDispatch, _ string, _ ...interface{}) (*ole.VARIANT, error) {
			return nil, callErr
		})
		_, err := (&oleClient{}).DirectoryHistories([]int{WindowLeft})
		if !errors.Is(err, callErr) {
			t.Fatalf("DirectoryHistories() error = %v, want wrapped %v", err, callErr)
		}
	})

	t.Run("unexpected count type", func(t *testing.T) {
		result := stringVariant("two")
		stubCallMethod(t, func(_ *ole.IDispatch, _ string, _ ...interface{}) (*ole.VARIANT, error) {
			return &result, nil
		})
		if _, err := (&oleClient{}).DirectoryHistories([]int{WindowLeft}); err == nil {
			t.Fatal("DirectoryHistories() error = nil")
		}
	})

	t.Run("history call", func(t *testing.T) {
		callErr := errors.New("history failed")
		calls := 0
		stubCallMethod(t, func(_ *ole.IDispatch, _ string, _ ...interface{}) (*ole.VARIANT, error) {
			calls++
			if calls == 1 {
				result := ole.NewVariant(ole.VT_I4, 1)
				return &result, nil
			}
			return nil, callErr
		})
		_, err := (&oleClient{}).DirectoryHistories([]int{WindowLeft})
		if !errors.Is(err, callErr) {
			t.Fatalf("DirectoryHistories() error = %v, want wrapped %v", err, callErr)
		}
	})
}

func TestCurrentFile(t *testing.T) {
	variables := []string{"$P", "$F"}
	values := []string{`C:\Work\`, "file.txt"}
	call := 0
	stubCallMethod(t, func(_ *ole.IDispatch, name string, params ...interface{}) (*ole.VARIANT, error) {
		if name != "Extract" || len(params) != 1 || params[0] != variables[call] {
			t.Fatalf("call %d = %s %#v", call, name, params)
		}
		result := stringVariant(values[call])
		call++
		return &result, nil
	})

	got, err := (&oleClient{}).CurrentFile()
	if err != nil {
		t.Fatalf("CurrentFile() error = %v", err)
	}
	if want := `C:\Work\file.txt`; got != want {
		t.Errorf("CurrentFile() = %q, want %q", got, want)
	}
}

func TestMarkedFilesUsesLineSeparatedMacro(t *testing.T) {
	result := stringVariant("C:\\My File.txt\nD:\\Other.txt")
	stubCallMethod(t, func(_ *ole.IDispatch, name string, params ...interface{}) (*ole.VARIANT, error) {
		if name != "Extract" || !reflect.DeepEqual(params, []interface{}{"$JU$QN$MF"}) {
			t.Fatalf("call = %s %#v", name, params)
		}
		return &result, nil
	})

	got, err := (&oleClient{}).MarkedFiles()
	if err != nil {
		t.Fatalf("MarkedFiles() error = %v", err)
	}
	want := []string{`C:\My File.txt`, `D:\Other.txt`}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MarkedFiles() = %v, want %v", got, want)
	}
}

func TestToInt(t *testing.T) {
	t.Run("supported types", func(t *testing.T) {
		tests := []struct {
			name  string
			input any
			want  int
		}{
			{"int", int(3), 3},
			{"int16", int16(3), 3},
			{"int32", int32(3), 3},
			{"int64", int64(3), 3},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := toInt(tt.input)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tt.want {
					t.Errorf("got %d, want %d", got, tt.want)
				}
			})
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		if _, err := toInt("3"); err == nil {
			t.Error("expected error for unsupported type, got nil")
		}
	})
}

func TestParseMarkedFiles(t *testing.T) {
	t.Run("marked files", func(t *testing.T) {
		got, err := parseMarkedFiles("C:\\a.txt\nC:\\b.txt", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{`C:\a.txt`, `C:\b.txt`}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("paths containing spaces", func(t *testing.T) {
		got, err := parseMarkedFiles("C:\\My Files\\a.txt\r\nD:\\Another Folder\\b.txt\n", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{`C:\My Files\a.txt`, `D:\Another Folder\b.txt`}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("no marks uses current file", func(t *testing.T) {
		got, err := parseMarkedFiles("  ", func() (string, error) {
			return `C:\current.txt`, nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(got, []string{`C:\current.txt`}) {
			t.Errorf("unexpected files: %v", got)
		}
	})

	t.Run("current file error", func(t *testing.T) {
		_, err := parseMarkedFiles("", func() (string, error) {
			return "", errors.New("current file failed")
		})
		if err == nil || err.Error() != "current file failed" {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
