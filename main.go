package dirsizecalc

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type NameSize struct {
	Name string
	Size float64
}

// Создаст и вернёт срез структур NameSize. Каждый элемент среза будет содержать в себе название папки корневой директории и её размер.
func GetDirectories(rootPath string) ([]NameSize, error) {
	//Создаём срез nameSizeArray (для хранения имени и размера папок)
	var nameSizeArray []NameSize

	//Проходим по всем файлам корневой директории.
	dirs, err := ioutil.ReadDir(rootPath)
	if err != nil {
		fmt.Println("Ошибка при чтении файлов ROOT директории:", err)
		return nameSizeArray, err
	}

	//Если очередной файл оказывается директорией, то запускается вычисление её размера
	//Размер и имя папки заносятся в ранее созданный nameSizeArray
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		//Вычисляем размер найденной директории
		c := make(chan float64) //Создаём канал, в который будут передаваться размеры найденных директорий
		defer close(c)
		go dirSizeCalculation(fmt.Sprintf("%s/%s", rootPath, dir.Name()), c)
		dirSize := <-c
		dirSizeMb := dirSize / (1024 * 1024)

		//Создаём переменную типа nameSize и добавления в nameSizeArray
		nameSizeValue := NameSize{dir.Name(), dirSizeMb}
		nameSizeArray = append(nameSizeArray, nameSizeValue)

		//Обработка возможной ошибки при возвращении в родительскую директорию
		err = os.Chdir("..")
		if err != nil {
			fmt.Println("..")
			return nameSizeArray, err
		}
	}

	return nameSizeArray, nil
}

// Вычисление размера директории с учётом вложенности в неё других директорий. Применяется многопоточное вычисление. Результат отправляется в заданный канал.
func dirSizeCalculation(path string, c chan<- float64) {
	//Открываем канал sizes для передачи в него размеров вложенных директорий
	sizes := make(chan int64)

	//Данная функция считает размер всех файлов, которые не являются директориями, и отправляет результат в канал sizes
	readSize := func(path string, file os.FileInfo, err error) error {
		if err != nil || file == nil {
			return nil
		}
		if !file.IsDir() {
			sizes <- file.Size()
		}
		return nil
	}

	//Проходим по всем вложенным директориям и вычисляем их размер.
	//После завершения работы всех горутин канал закрывается.
	go func() {
		filepath.Walk(path, readSize)
		close(sizes)
	}()

	//Суммируем все данных, которые находятся в канале sizes и получаем итоговый размер директории
	size := int64(0)
	for s := range sizes {
		size += s
	}

	//Возвращаем размер
	c <- float64(size)
}
