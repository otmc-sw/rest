/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package handlers

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	rest "github.com/otmc-sw/rest"
)

func generateDownloadFilePath(id int64) string {
	log.Println("Generating download file path for ID:", id)
	return fmt.Sprintf("/test/download_%d.txt", id)
}

func generateUploadDir(id int64) string {
	log.Println("Generating upload directory for ID:", id)
	return "/test"
}

func DownloadFile(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, _ := strconv.ParseInt(id, 10, 64)
	filePath := generateDownloadFilePath(idInt)

	return rest.
		Download(FiberContext{Ctx: c}).
		Source("./data/files" + filePath).
		Respond()
}

func UploadFile(c *fiber.Ctx) error {
	var file rest.File
	parentId := c.Query("parent_id")
	parentIdInt, _ := strconv.ParseInt(parentId, 10, 64)
	parentPath := generateUploadDir(parentIdInt)
	fullPath := "./data/files" + parentPath + "/" + file.Name

	return rest.
		Upload(FiberContext{Ctx: c}).
		Destination(fullPath).
		Bind(&file).
		Respond()
}

func ReadFileContent(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, _ := strconv.ParseInt(id, 10, 64)
	filePath := generateDownloadFilePath(idInt)

	return rest.
		ReadFileContent(FiberContext{Ctx: c}).
		Source("./data/files" + filePath).
		Respond()
}

func UpdateFileContent(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, _ := strconv.ParseInt(id, 10, 64)
	filePath := generateDownloadFilePath(idInt)

	return rest.
		UpdateFileContent(FiberContext{Ctx: c}).
		Source("./data/files" + filePath).
		Respond()
}
