package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {

	return walkDir(out, path, printFiles, "")
}

func walkDir(out io.Writer, path string, printFiles bool, prefix string) error {
	// Получаем содержимое текущей директории
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// Фильтруем элементы исходя из аргуменитов
	filtered := []fs.DirEntry{}
	for _, e := range entries {
		if e.IsDir() || printFiles {
			filtered = append(filtered, e)
		}
	}

	// Сортируем элементы по имени
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name() < filtered[j].Name()
	})

	// Обходим каждый элемент в текущей директории
	for i, entry := range filtered {

		last := i == len(filtered)-1

		line := "├───"

		if last {
			line = "└───"
		}
		line += entry.Name()

		// Если файл, то показываем его размер
		if !entry.IsDir() && printFiles {
			info, _ := entry.Info()
			size := info.Size()

			if size == 0 {
				line += " (empty)"
			} else {
				line += fmt.Sprintf(" (%vb)", size)
			}
		}

		// Выводим текущую строку с отступами
		fmt.Fprintf(out, "%s%s\n", prefix, line)

		// Если директория, то рекурсивно обходим ее
		if entry.IsDir() {
			newPrefix := prefix

			if last {
				newPrefix += "\t"
			} else {
				newPrefix += "│\t"
			}

			fullPath := filepath.Join(path, entry.Name())

			walkDir(out, fullPath, printFiles, newPrefix)
		}
	}

	return nil
}
