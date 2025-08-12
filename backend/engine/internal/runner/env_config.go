package runner

import (
	"pariksha/common/pkg/constants"
	engineConstants "pariksha/engine/internal/constants"
	"pariksha/engine/internal/templates"
)

var envConfigs = map[string]environmentConfig{
	constants.LangNode: {
		Image:        engineConstants.NodeImage,
		FileExt:      ".js",
		CommandName:  "node",
		CommandArgs:  nil,
		MountTarget:  "/code/solution.js",
		TemplateFunc: templates.GenerateNodeScript,
	},
}
