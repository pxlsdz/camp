drop database camp;

CREATE DATABASE IF NOT EXISTS camp
    DEFAULT CHARACTER SET utf8
    DEFAULT COLLATE utf8_general_ci;

use camp;

create table member
(
    id        bigint             not null auto_increment primary key comment 'id',
    username  varchar(20) binary not null unique comment '用户名',
    password  varchar(20) binary not null comment '密码',
    nickname  varchar(20) binary not null comment '用户昵称',
    user_type tinyint            not null comment '用户类型 1:管理员 2:学生 3:教师',
    deleted   tinyint            not null default 0 comment '删除 0:未删除 1:已删除'
) comment ='用户';

create table course
(
    id         bigint              not null auto_increment primary key comment 'id',
    name       varchar(255) binary not null unique comment '课程名称',
    cap        int                 not null comment '课程容量',
    teacher_id bigint                       default null comment '老师id',
    deleted    tinyint             not null default 0 comment '删除 0:未删除 1:已删除'
) comment ='课程表';


CREATE TABLE student_course
(
    `id`         bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
    `student_id` bigint NOT NULL COMMENT '学生id',
    `course_id`  bigint NOT NULL COMMENT '课程id',
    PRIMARY KEY (`id`),
    UNIQUE KEY `sc_index` (`student_id`, `course_id`) USING BTREE
) COMMENT ='学生课程关系表';


INSERT INTO member (username, password, nickname, user_type)
VALUES ('JudgeAdmin', 'JudgePassword2022', '宋端正', 1),
       ('sdzsdz', 'JudgePassword2022', 'sdzsdz', 2);

# INSERT INTO course (name, cap)
# VALUES ('高数', 1000),
#        ('英语', 500);

truncate course