package utils

// 常见的工具类型列表
var utilityTypes = map[string]bool{
	"Omit": true, "Pick": true, "Partial": true, "Required": true,
	"Readonly": true, "Record": true, "Exclude": true, "Extract": true,
	"NonNullable": true, "Parameters": true, "ReturnType": true,
	"InstanceType": true, "ThisParameterType": true, "OmitThisParameter": true,
	"ThisType": true, "Uppercase": true, "Lowercase": true,
	"Capitalize": true, "Uncapitalize": true,
}

// IsUtilityType 检查给定的类型是否是工具类型
func IsUtilityType(typeName string) bool {
	_, exists := utilityTypes[typeName]
	return exists
}
