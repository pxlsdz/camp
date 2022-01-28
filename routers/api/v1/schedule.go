package v1

import (
	"bufio"
	"camp/types"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type scheduleParameter struct {
	teacher2course *map[string][]string
	course2teacher *map[string][]string
	plan4teacher   *map[string]string
	plan4course    *map[string]string
	vis4teacher    *map[string]bool
	vis4course     *map[string]bool
}

func hungry(x string, param *scheduleParameter, mark int) int {
	if mark == 0 {
		for i := 0; i < len((*param.teacher2course)[x]); i++ {
			course := (*param.teacher2course)[x][i]
			if (*param.vis4course)[course] == false {
				(*param.vis4course)[course] = true
				if _, ok := (*param.plan4course)[course]; ok == false {
					(*param.plan4teacher)[x] = course
					(*param.plan4course)[course] = x
					// fmt.Println("#1 Change", x, course)
					return 1
				} else if hungry((*param.plan4course)[course], param, 0) == 1 {
					(*param.plan4teacher)[x] = course
					(*param.plan4course)[course] = x
					// fmt.Println("#1 Change", x, course)
					return 1
				}
			}
		}
	} else if mark == 1 {
		for i := 0; i < len((*param.course2teacher)[x]); i++ {
			teacher := (*param.course2teacher)[x][i]
			if (*param.vis4teacher)[teacher] == false {
				(*param.vis4teacher)[teacher] = true
				if _, ok := (*param.plan4teacher)[teacher]; ok == false {
					(*param.plan4teacher)[teacher] = x
					(*param.plan4course)[x] = teacher
					// fmt.Println("#2 Change", teacher, x)
					return 1
				} else if hungry((*param.plan4teacher)[teacher], param, 1) == 1 {
					(*param.plan4teacher)[teacher] = x
					(*param.plan4course)[x] = teacher
					// fmt.Println("#2 Change", teacher, x)
					return 1
				}
			}
		}

	} else {
		fmt.Println("unknown error")
	}

	return 0
}

func ScheduleCore(TeacherCourseRelationShip *map[string][]string) (int, map[string]string) {
	course2teacher := make(map[string][]string)
	teacherList := make([]string, 0)
	for k, v := range *TeacherCourseRelationShip {
		for i := 0; i < len(v); i++ {
			course2teacher[v[i]] = append(course2teacher[v[i]], k)
		}
		teacherList = append(teacherList, k)
	}

	courseList := make([]string, 0)
	for k, _ := range course2teacher {
		courseList = append(courseList, k)
	}

	sum := 0
	plan4teacher := make(map[string]string, len(teacherList))
	plan4course := make(map[string]string, len(courseList))
	for i := 0; i < len(teacherList); i++ {

		vis4teacher := make(map[string]bool, len(teacherList))
		vis4course := make(map[string]bool, len(courseList))

		sum += hungry(teacherList[i], &scheduleParameter{
			teacher2course: TeacherCourseRelationShip,
			course2teacher: &course2teacher,
			plan4teacher:   &plan4teacher,
			plan4course:    &plan4course,
			vis4teacher:    &vis4teacher,
			vis4course:     &vis4course,
		}, 0)
	}
	return sum, plan4teacher
}
func ScheduleCourse(c *gin.Context) {
	//TODO:登录验证和权限认证

	var json types.ScheduleCourseRequest

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, types.ScheduleCourseResponse{Code: types.ParamInvalid})
	}
	_, ret := ScheduleCore(&json.TeacherCourseRelationShip)
	c.JSON(http.StatusOK, types.ScheduleCourseResponse{
		Code: types.OK,
		Data: ret,
	})
}
func ReadRequest(path string) map[string][]string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("打开文件出错：%v\n", err)
	}
	defer file.Close()
	// bufio.NewReader(rd io.Reader) *Reader
	reader := bufio.NewReader(file)
	// 读取n m

	line, err := reader.ReadString('\n') // 读到一个换行符就结束
	words := strings.Split(line, " ")
	words[1] = strings.Split(words[1], "\n")[0]

	n, _ := strconv.Atoi(words[0])
	m, _ := strconv.Atoi(words[1])
	fmt.Println("n=", n, "m=", m)

	// 读取 relation
	var relation map[string][]string
	relation = make(map[string][]string, n)
	for i := 1; i <= n; i++ {
		line, err = reader.ReadString('\n') // 读到一个换行符就结束
		words = strings.Split(line, " ")
		words[len(words)-1] = strings.Split(words[len(words)-1], "\n")[0]
		for j := 1; j < len(words); j++ {
			relation[strconv.Itoa(i)] = append(relation[strconv.Itoa(i)], words[j])
		}
	}
	return relation
}

func ScheduleCourseTest(c *gin.Context) {
	c.JSON(http.StatusOK,
		types.ScheduleCourseRequest{TeacherCourseRelationShip: ReadRequest("testdata/input6.txt")})
}
