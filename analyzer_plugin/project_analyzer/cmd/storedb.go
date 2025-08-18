package cmd

// example: go run main.go store-db -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result/analyzer_result_db -x "node_modules/**" -x "bffApiDoc/**"

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"main/analyzer/projectParser"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

func NewStoreDbCmd() *cobra.Command {
	storeDbCmd := &cobra.Command{
		Use:   "store-db",
		Short: "分析 TypeScript 项目并将结果存储在 SQLite 数据库中。",
		Long:  `该命令分析给定的 TypeScript 项目，将 package.json、代码文件和各类代码节点（导入、导出、接口等）分别存入其专用的数据库宽表中。`,
		Run: func(cmd *cobra.Command, args []string) {
			inputPath, _ := cmd.Flags().GetString("input")
			outputDir, _ := cmd.Flags().GetString("output") // 现在是输出目录
			excludePatterns, _ := cmd.Flags().GetStringSlice("exclude")
			isMonorepo, _ := cmd.Flags().GetBool("monorepo")

			if inputPath == "" || outputDir == "" {
				log.Fatal("需要提供输入和输出路径。")
			}

			// 根据输入目录名，自动生成数据库文件名
			projectName := filepath.Base(inputPath)
			dbFileName := fmt.Sprintf("%s.db", projectName)
			finalDbPath := filepath.Join(outputDir, dbFileName)

			// 确保输出目录存在
			if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
				log.Fatalf("无法创建输出目录 %s: %v", outputDir, err)
			}

			fmt.Println("开始分析...")
			config := projectParser.NewProjectParserConfig(inputPath, excludePatterns, isMonorepo)
			projectData := projectParser.NewProjectParserResult(config)
			projectData.ProjectParser()
			fmt.Println(fmt.Sprintf("分析完成。发现 %d 个JS/TS文件和 %d 个package.json文件。", len(projectData.Js_Data), len(projectData.Package_Data)))

			fmt.Println("正在将结果存储到数据库:", finalDbPath)
			if err := storeInDatabase(projectData, finalDbPath); err != nil {
				log.Fatalf("无法将数据存储到数据库: %v", err)
			}
			fmt.Println("成功将分析结果存储在", finalDbPath)
		},
	}

	storeDbCmd.Flags().StringP("input", "i", "", "要分析的 TypeScript 项目目录的路径")
	// 更新 output 标志的描述
	storeDbCmd.Flags().StringP("output", "o", "", "用于存储数据库文件的输出目录路径")
	storeDbCmd.Flags().StringSliceP("exclude", "x", []string{}, "要从分析中排除的 Glob 模式 (可多次指定)")
	storeDbCmd.Flags().BoolP("monorepo", "m", false, "如果要分析的是 monorepo，则设置为 true")
	storeDbCmd.MarkFlagRequired("input")
	storeDbCmd.MarkFlagRequired("output")

	return storeDbCmd
}

// storeInDatabase 负责将所有分析数据存入数据库
func storeInDatabase(projectData *projectParser.ProjectParserResult, dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("无法打开数据库: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("无法连接到数据库: %w", err)
	}

	if err := createSchema(db); err != nil {
		return fmt.Errorf("无法创建数据库结构: %w", err)
	}

	projectID, err := generateProjectID()
	if err != nil {
		return fmt.Errorf("无法生成项目ID: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("无法开始事务: %w", err)
	}

	// 准备所有表的插入语句
	stmtCache := make(map[string]*sql.Stmt)
	queries := map[string]string{
		"packages":         "INSERT INTO packages (project_id, path, workspace, namespace, version) VALUES (?, ?, ?, ?, ?)",
		"npm_dependencies": "INSERT INTO npm_dependencies (package_id, name, dependency_type, declared_version, installed_version) VALUES (?, ?, ?, ?, ?)",
		"files":            "INSERT INTO files (project_id, path) VALUES (?, ?)",
		"imports":          "INSERT INTO imports (file_id, identifier, kind, source_path, raw_code) VALUES (?, ?, ?, ?, ?)",
		"exports":          "INSERT INTO exports (file_id, identifier, kind, source_path, raw_code) VALUES (?, ?, ?, ?, ?)",
		"interfaces":       "INSERT INTO interfaces (file_id, identifier, line_start, line_end, references_json, raw_code) VALUES (?, ?, ?, ?, ?, ?)",
		"function_calls":   "INSERT INTO function_calls (file_id, call_chain, line_start, line_end, arguments_json, raw_code) VALUES (?, ?, ?, ?, ?, ?)",
		"variables":        "INSERT INTO variables (file_id, identifier, declaration_kind, exported, line_start, line_end, details_json, raw_code) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
	}

	for name, query := range queries {
		stmt, err := tx.Prepare(query)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("无法准备 '%s' 表的插入语句: %w", name, err)
		}
		stmtCache[name] = stmt
		defer stmt.Close()
	}

	// 1. 存储 Package 数据
	for path, pkgData := range projectData.Package_Data {
		res, err := stmtCache["packages"].Exec(projectID, path, pkgData.Workspace, pkgData.Namespace, pkgData.Version)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("无法插入 package %s: %w", path, err)
		}
		packageID, _ := res.LastInsertId()

		for _, npmItem := range pkgData.NpmList {
			_, err := stmtCache["npm_dependencies"].Exec(packageID, npmItem.Name, npmItem.Type, npmItem.Version, npmItem.NodeModuleVersion)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("无法插入 npm 依赖 %s: %w", npmItem.Name, err)
			}
		}
	}

	// 2. 存储 JS/TS 文件及代码节点数据
	for path, jsFileData := range projectData.Js_Data {
		res, err := stmtCache["files"].Exec(projectID, path)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("无法插入文件 %s: %w", path, err)
		}
		fileID, _ := res.LastInsertId()

		// 存储导入
		for _, imp := range jsFileData.ImportDeclarations {
			for _, mod := range imp.ImportModules {
				_, err := stmtCache["imports"].Exec(fileID, mod.Identifier, mod.Type, imp.Source.FilePath, imp.Raw)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("无法在文件 %s 中插入导入: %w", path, err)
				}
			}
		}

		// 存储导出
		for _, exp := range jsFileData.ExportDeclarations {
			for _, mod := range exp.ExportModules {
				sourcePath := ""
				if exp.Source != nil {
					sourcePath = exp.Source.FilePath
				}
				_, err := stmtCache["exports"].Exec(fileID, mod.Identifier, mod.Type, sourcePath, exp.Raw)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("无法在文件 %s 中插入导出: %w", path, err)
				}
			}
		}

		// 存储接口
		for name, iface := range jsFileData.InterfaceDeclarations {
			referencesJSON, _ := json.Marshal(iface.Reference)
			_, err := stmtCache["interfaces"].Exec(fileID, name, iface.SourceLocation.Start.Line, iface.SourceLocation.End.Line, string(referencesJSON), iface.Raw)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("无法在文件 %s 中插入接口 %s: %w", path, name, err)
			}
		}

		// 存储函数调用
		for _, call := range jsFileData.CallExpressions {
			argumentsJSON, _ := json.Marshal(call.Arguments)
			_, err := stmtCache["function_calls"].Exec(fileID, strings.Join(call.CallChain, "."), call.SourceLocation.Start.Line, call.SourceLocation.End.Line, string(argumentsJSON), call.Raw)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("无法在文件 %s 中插入函数调用: %w", path, err)
			}
		}

		// 存储变量
		for _, v := range jsFileData.VariableDeclarations {
			for _, declarator := range v.Declarators {
				detailsJSON, _ := json.Marshal(declarator)
				_, err := stmtCache["variables"].Exec(fileID, declarator.Identifier, string(v.Kind), v.Exported, v.SourceLocation.Start.Line, v.SourceLocation.End.Line, string(detailsJSON), v.Raw)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("无法在文件 %s 中插入变量 %s: %w", path, declarator.Identifier, err)
				}
			}
		}
	}

	return tx.Commit()
}

// createSchema 负责创建所有专用的宽表
func createSchema(db *sql.DB) error {
	schema := `
    DROP TABLE IF EXISTS npm_dependencies;
    DROP TABLE IF EXISTS packages;
    DROP TABLE IF EXISTS imports;
    DROP TABLE IF EXISTS exports;
    DROP TABLE IF EXISTS interfaces;
    DROP TABLE IF EXISTS function_calls;
    DROP TABLE IF EXISTS variables;
    DROP TABLE IF EXISTS files;

    CREATE TABLE packages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        project_id TEXT NOT NULL,
        path TEXT NOT NULL UNIQUE,
        workspace TEXT,
        namespace TEXT,
        version TEXT
    );

    CREATE TABLE npm_dependencies (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        package_id INTEGER NOT NULL,
        name TEXT NOT NULL,
        dependency_type TEXT NOT NULL,
        declared_version TEXT,
        installed_version TEXT,
        FOREIGN KEY (package_id) REFERENCES packages (id)
    );

    CREATE TABLE files (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        project_id TEXT NOT NULL,
        path TEXT NOT NULL UNIQUE
    );

    CREATE TABLE imports (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        file_id INTEGER NOT NULL,
        identifier TEXT,
        kind TEXT,
        source_path TEXT,
        raw_code TEXT,
        FOREIGN KEY (file_id) REFERENCES files (id)
    );

    CREATE TABLE exports (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        file_id INTEGER NOT NULL,
        identifier TEXT,
        kind TEXT,
        source_path TEXT,
        raw_code TEXT,
        FOREIGN KEY (file_id) REFERENCES files (id)
    );

    CREATE TABLE interfaces (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        file_id INTEGER NOT NULL,
        identifier TEXT,
        line_start INTEGER,
        line_end INTEGER,
        references_json TEXT,
        raw_code TEXT,
        FOREIGN KEY (file_id) REFERENCES files (id)
    );

    CREATE TABLE function_calls (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        file_id INTEGER NOT NULL,
        call_chain TEXT,
        line_start INTEGER,
        line_end INTEGER,
        arguments_json TEXT,
        raw_code TEXT,
        FOREIGN KEY (file_id) REFERENCES files (id)
    );

    CREATE TABLE variables (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        file_id INTEGER NOT NULL,
        identifier TEXT,
        declaration_kind TEXT,
        exported BOOLEAN,
        line_start INTEGER,
        line_end INTEGER,
        details_json TEXT,
        raw_code TEXT,
        FOREIGN KEY (file_id) REFERENCES files (id)
    );
    `
	_, err := db.Exec(schema)
	return err
}

// generateProjectID 生成一个随机的、唯一的字符串用作项目ID。
func generateProjectID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
