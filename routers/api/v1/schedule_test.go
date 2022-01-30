package v1

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestScheduleCore(t *testing.T) {
	type tCase struct {
		TeacherCourseRelationShip map[string][]string
		expected                  int
	}
	cases := make([]*tCase, 0, 0)
	for i := 1; i <= 9; i++ {
		cases = append(cases, &tCase{
			TeacherCourseRelationShip: ReadRequest("../../../testdata/input" + strconv.Itoa(i) + ".txt"),
			expected:                  ReadResult("../../../testdata/output" + strconv.Itoa(i) + ".txt"),
		})
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("case%d", i), func(t *testing.T) {
			got, _ := ScheduleCore(&c.TeacherCourseRelationShip)
			if got != c.expected {
				t.Errorf("Case %d expected %d, but %d got", i, c.expected, got)
			}
		})
	}
}
func ReadResult(path string) int {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("打开文件出错：%v\n", err)
	}
	defer file.Close()
	// bufio.NewReader(rd io.Reader) *Reader
	reader := bufio.NewReader(file)
	// 读取答案
	line, err := reader.ReadString('\n') // 读到一个换行符就结束
	ans := strings.Split(line, "\n")[0]
	res, _ := strconv.Atoi(ans)
	return res
}
