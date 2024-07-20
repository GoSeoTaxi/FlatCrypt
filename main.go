package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	maxFileNameLength = 255
	pathSeparator     = "_!_"
	illegalChars      = "<>:\"/\\|?*"
)

type FileInfo struct {
	Path     string
	IsDir    bool
	FileInfo os.FileInfo
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: program <encrypt/decrypt> <source_dir> <dest_dir>")
		return
	}

	mode := os.Args[1]
	sourceDir := os.Args[2]
	destDir := os.Args[3]

	// Проверка существования исходной директории
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		fmt.Printf("Error: Source directory '%s' does not exist\n", sourceDir)
		return
	}

	// Проверка существования директории назначения, создание если не существует
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		fmt.Printf("Destination directory '%s' does not exist. Creating it...\n", destDir)
		err = os.MkdirAll(destDir, 0755)
		if err != nil {
			fmt.Printf("Error creating destination directory: %v\n", err)
			return
		}
		fmt.Println("Destination directory created successfully.")
	}

	switch mode {
	case "encrypt":
		if err := validateEncryption(sourceDir); err != nil {
			fmt.Printf("Validation failed: %v\n", err)
			return
		}
		err := encryptDirectory(sourceDir, destDir)
		if err != nil {
			fmt.Printf("Error encrypting: %v\n", err)
		}
	case "decrypt":
		err := decryptDirectory(sourceDir, destDir)
		if err != nil {
			fmt.Printf("Error decrypting: %v\n", err)
		}
	default:
		fmt.Println("Invalid mode. Use 'encrypt' or 'decrypt'.")
	}
}

func validateEncryption(sourceDir string) error {
	var longPaths []string

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		encryptedName := encryptPath(relPath, 1) // Используем 1 как заглушку для номера
		if len(encryptedName) > maxFileNameLength {
			longPaths = append(longPaths, relPath)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking the path %s: %v", sourceDir, err)
	}

	if len(longPaths) > 0 {
		return fmt.Errorf("the following paths exceed %d characters when encrypted:\n%s",
			maxFileNameLength, strings.Join(longPaths, "\n"))
	}

	return nil
}

func encryptDirectory(sourceDir, destDir string) error {
	var files []FileInfo

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		files = append(files, FileInfo{Path: relPath, IsDir: info.IsDir(), FileInfo: info})
		return nil
	})

	if err != nil {
		return err
	}

	sort.Slice(files, func(i, j int) bool {
		// Файлы корневого каталога идут первыми
		iIsRoot := filepath.Dir(files[i].Path) == "."
		jIsRoot := filepath.Dir(files[j].Path) == "."
		if iIsRoot != jIsRoot {
			return iIsRoot
		}

		// Затем сортируем по директории
		iDir := filepath.Dir(files[i].Path)
		jDir := filepath.Dir(files[j].Path)
		if iDir != jDir {
			return iDir < jDir
		}

		// Если директории одинаковые, сортируем файлы перед директориями
		if files[i].IsDir != files[j].IsDir {
			return !files[i].IsDir
		}

		// Если оба элемента - файлы или директории, сортируем по имени
		return files[i].Path < files[j].Path
	})

	globalCounter := 0

	for _, file := range files {
		if file.IsDir {
			continue
		}

		globalCounter++
		encryptedName := encryptPath(file.Path, globalCounter)
		destPath := filepath.Join(destDir, encryptedName)

		err := copyFile(filepath.Join(sourceDir, file.Path), destPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func decryptDirectory(sourceDir, destDir string) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		decryptedPath := decryptPath(info.Name())
		fullDestPath := filepath.Join(destDir, decryptedPath)

		err = os.MkdirAll(filepath.Dir(fullDestPath), 0755)
		if err != nil {
			return err
		}

		return copyFile(path, fullDestPath)
	})
}

func encryptPath(path string, fileNumber int) string {
	// Заменяем все разделители пути на наш собственный разделитель
	encryptedName := strings.ReplaceAll(filepath.ToSlash(path), "/", pathSeparator)

	// Заменяем недопустимые символы на подчеркивание
	for _, char := range illegalChars {
		encryptedName = strings.ReplaceAll(encryptedName, string(char), "_")
	}

	// Добавляем порядковый номер с ведущими нулями
	encryptedName = fmt.Sprintf("%03d%s%s", fileNumber, pathSeparator, encryptedName)

	// Ограничиваем длину имени файла
	if len(encryptedName) > maxFileNameLength {
		encryptedName = encryptedName[:maxFileNameLength]
	}

	return encryptedName
}

func decryptPath(encryptedName string) string {
	// Удаляем порядковый номер и разделитель
	parts := strings.SplitN(encryptedName, pathSeparator, 3)
	if len(parts) < 3 {
		return encryptedName
	}
	encryptedName = parts[2]

	// Заменяем наш разделитель обратно на разделитель пути
	return filepath.FromSlash(strings.ReplaceAll(encryptedName, pathSeparator, "/"))
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
