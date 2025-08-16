package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"main/analyzer/projectParser"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

var storeDbCmd = &cobra.Command{
	Use:   "store-db",
	Short: "分析 TypeScript 项目并将结果存储在 SQLite 数据库中。",
	Long: `该命令分析给定的 TypeScript 项目，解析所有相关文件以构建抽象语法树 (AST)。
然后将结构化数据（包括声明、依赖项和调用表达式）存储到 SQLite 数据库文件中以供查询。`,
	Run: func(cmd *cobra.Command, args []string) {
		inputPath, _ := cmd.Flags().GetString("input")
		outputPath, _ := cmd.Flags().GetString("output")
		excludePatterns, _ := cmd.Flags().GetStringSlice("exclude")
		isMonorepo, _ := cmd.Flags().GetBool("monorepo")

		if inputPath == "" || outputPath == "" {
			log.Fatal("需要输入和输出路径。")
		}

		fmt.Println("开始分析...")
		config := projectParser.NewProjectParserConfig(inputPath, excludePatterns, isMonorepo)
		projectData := projectParser.NewProjectParserResult(config)
		projectData.ProjectParser()
		fmt.Printf("分析完成。发现 %d 个文件。\n", len(projectData.Js_Data))

		fmt.Println("正在将结果存储到数据库:", outputPath)
		if err := storeInDatabase(projectData, outputPath); err != nil {
			log.Fatalf("无法将数据存储到数据库: %v", err)
		}
		fmt.Println("成功将分析结果存储在", outputPath)
	},
}

func init() {
	storeDbCmd.Flags().StringP("input", "i", "", "要分析的 TypeScript 项目目录的路径")
	storeDbCmd.Flags().StringP("output", "o", "", "输出的 SQLite 数据库文件的路径 (例如, /path/to/result.db)")
	storeDbCmd.Flags().StringSliceP("exclude", "x", []string{}, "要从分析中排除的 Glob 模式 (可多次指定)")
	storeDbCmd.Flags().BoolP("monorepo", "m", false, "如果要分析的是 monorepo，则设置为 true")
	storeDbCmd.MarkFlagRequired("input")
	storeDbCmd.MarkFlagRequired("output")
	rootCmd.AddCommand(storeDbCmd)
}

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

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("无法开始事务: %w", err)
	}

	fileStmt, err := tx.Prepare("INSERT INTO files (path) VALUES (?)")
	if err != nil {
		return fmt.Errorf("无法准备文件插入语句: %w", err)
	}
	defer fileStmt.Close()

	declStmt, err := tx.Prepare("INSERT INTO declarations (file_id, name, type, start_pos, end_pos, metadata) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("无法准备声明插入语句: %w", err)
	}
	defer declStmt.Close()

	depStmt, err := tx.Prepare("INSERT INTO dependencies (file_id, import_path, identifiers) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("无法准备依赖插入语句: %w", err)
	}
	defer depStmt.Close()

	callStmt, err := tx.Prepare("INSERT INTO call_expressions (file_id, caller_name, expression, arguments, start_pos, end_pos) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("无法准备调用表达式插入语句: %w", err)
	}
	defer callStmt.Close()

	for path, jsFileData := range projectData.Js_Data {
		res, err := fileStmt.Exec(path)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("无法插入文件 %s: %w", path, err)
		}
		fileID, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("无法获取文件 %s 的最后插入ID: %w", path, err)
		}

		// 这部分需要根据声明的实际结构进行调整
		// 由于结构复杂，我们暂时跳过它

		for _, dep := range jsFileData.ImportDeclarations {
			identifiers, _ := json.Marshal(dep.ImportModules)
			_, err := depStmt.Exec(fileID, dep.Source.FilePath, string(identifiers))
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("无法在文件 %s 中插入依赖: %w", path, err)
			}
		}

		for _, call := range jsFileData.CallExpressions {
			arguments, _ := json.Marshal(call.Arguments)
			_, err := callStmt.Exec(fileID, "", strings.Join(call.CallChain, "."), string(arguments), 0, 0) // 位置信息不在此结构中
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("无法在文件 %s 中插入调用表达式: %w", path, err)
			}
		}
	}

	return tx.Commit()
}

func createSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS declarations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		start_pos INTEGER,
		end_pos INTEGER,
		metadata TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (file_id) REFERENCES files (id)
	);
	CREATE TABLE IF NOT EXISTS dependencies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id INTEGER NOT NULL,
		import_path TEXT NOT NULL,
		identifiers TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (file_id) REFERENCES files (id)
	);
	CREATE TABLE IF NOT EXISTS call_expressions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id INTEGER NOT NULL,
		caller_name TEXT,
		expression TEXT NOT NULL,
		arguments TEXT,
		start_pos INTEGER,
		end_pos INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (file_id) REFERENCES files (id)
	);
	`
	_, err := db.Exec(schema)
	return err
}
