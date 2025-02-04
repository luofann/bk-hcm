# version
PRO_DIR   = $(shell pwd)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
VERSION   = $(shell echo ${ENV_BK_HCM_VERSION})
DEBUG     = $(shell echo ${ENV_BK_HCM_ENABLE_DEBUG})

# output directory for release package and version for command line
ifeq ("$(VERSION)", "")
	export OUTPUT_DIR = ${PRO_DIR}/build/bk-hcm
	export LDVersionFLAG = "-X hcm/pkg/version.BUILDTIME=${BUILDTIME} \
		-X hcm/pkg/version.DEBUG=${DEBUG}"
else
	GITHASH   = $(shell git rev-parse HEAD)
	export OUTPUT_DIR = ${PRO_DIR}/build/bk-hcm-${VERSION}
	export LDVersionFLAG = "-X hcm/pkg/version.VERSION=${VERSION} \
    	-X hcm/pkg/version.BUILDTIME=${BUILDTIME} \
    	-X hcm/pkg/version.GITHASH=${GITHASH} \
    	-X hcm/pkg/version.DEBUG=${DEBUG}"
endif

export GO111MODULE=on

include ./scripts/makefile/uname.mk

default: all

# 创建编译文件存储目录
pre:
	@echo -e "\e[34;1mBuilding...\n\033[0m"
	mkdir -p ${OUTPUT_DIR}

# 本地测试前后端编译
all: pre ui server
	@cd ${PRO_DIR}/cmd && make
	@echo -e "\e[34;1mBuild All Success!\n\033[0m"

# 后端本地测试编译
server: pre
	@cd ${PRO_DIR}/cmd && make
	@echo -e "\e[34;1mBuild Server Success!\n\033[0m"

# 二进制出包编译
package: pre ui api ver
	@echo -e "\e[34;1mPackaging...\n\033[0m"
	@mkdir -p ${OUTPUT_DIR}/bin
	@mkdir -p ${OUTPUT_DIR}/etc
	@mkdir -p ${OUTPUT_DIR}/install
	@mkdir -p ${OUTPUT_DIR}/install/sql
	@cp -f ${PRO_DIR}/scripts/install/migrate.sh ${OUTPUT_DIR}/install/
	@cp -rf ${PRO_DIR}/scripts/sql/* ${OUTPUT_DIR}/install/sql/
	@cd ${PRO_DIR}/cmd && make package
	@echo -e "\e[34;1mPackage All Success!\n\033[0m"

# 容器化编译
docker: pre ui ver
	@echo -e "\e[34;1mMake Dockering...\n\033[0m"
	@cp -rf ${PRO_DIR}/docs/support-file/docker/* ${OUTPUT_DIR}/
	@mv ${OUTPUT_DIR}/front ${OUTPUT_DIR}/bk-hcm-webserver/
	@cp -rf ${PRO_DIR}/scripts/sql ${OUTPUT_DIR}/bk-hcm-dataservice/
	@cd ${PRO_DIR}/cmd && make docker
	@echo -e "\e[34;1mMake Docker All Success!\n\033[0m"

# 编译前端
ui: pre
	@echo -e "\e[34;1mBuilding Front...\033[0m"
	@cd ${PRO_DIR}/front && npm i && npm run build
	@mv ${PRO_DIR}/front/dist ${OUTPUT_DIR}/front
	@echo -e "\e[34;1mBuild Front Success!\n\033[0m"

# 添加Api文档到编译文件中
api: pre
	@echo -e "\e[34;1mPackaging API Docs...\033[0m"
	@mkdir -p ${OUTPUT_DIR}/api/
	@mkdir -p ${OUTPUT_DIR}/api/api-server
	@cp -f docs/api-docs/api-server/api/bk_apigw_resources_bk-hcm.yaml ${OUTPUT_DIR}/api/api-server
	@tar -czf ${OUTPUT_DIR}/api/api-server/zh.tgz -C docs/api-docs/api-server/docs zh
	@echo -e "\e[34;1mPackaging API Docs Done\n\033[0m"

# 添加版本信息到编译文件中
ver: pre
	@echo ${VERSION} > ${OUTPUT_DIR}/VERSION
	@cp -rf ${PRO_DIR}/CHANGELOG.md ${OUTPUT_DIR}

# 清理编译文件
clean:
	@cd ${PRO_DIR}/cmd && make clean
	@rm -rf ${PRO_DIR}/build

# 初始化下载项目开发依赖工具
init-tools:
	# 前端代码检查依赖工具下载
	curl -o- -L https://yarnpkg.com/install.sh | bash
