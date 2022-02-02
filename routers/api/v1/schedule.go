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
	vis4course     *map[string]bool
}

func hungry(x string, param *scheduleParameter) int {
	for i := 0; i < len((*param.teacher2course)[x]); i++ {
		course := (*param.teacher2course)[x][i]
		if (*param.vis4course)[course] == false {
			(*param.vis4course)[course] = true
			if _, ok := (*param.plan4course)[course]; ok == false {
				(*param.plan4teacher)[x] = course
				(*param.plan4course)[course] = x
				return 1
			} else if hungry((*param.plan4course)[course], param) == 1 {
				(*param.plan4teacher)[x] = course
				(*param.plan4course)[course] = x
				return 1
			}
		}
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
		vis4course := make(map[string]bool, len(courseList))
		sum += hungry(teacherList[i], &scheduleParameter{
			teacher2course: TeacherCourseRelationShip,
			course2teacher: &course2teacher,
			plan4teacher:   &plan4teacher,
			plan4course:    &plan4course,
			vis4course:     &vis4course,
		})
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

	defer c.JSON(http.StatusOK, types.ScheduleCourseResponse{Code: types.UnknownError})
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
	// m, _ := strconv.Atoi(words[1])
	// fmt.Println("n=", n, "m=", m)

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
