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

// Создаст и вернёт два среза структур NameSize. Элементы срезов содержат в себе название файла корневой директории и его размер.
// Первый срез содержит в себе поддиректории корневой директории, а второй - файлы корневой директории.
func GetDirectories(rootPath string) ([]NameSize, []NameSize, error) {
	//Создаём срез nameSizeArray (для хранения имени и размера папок)
	var nameSizeArray []NameSize

	//Создаём срез nameSizeArray (для хранения имени и размера файлов)
	var filesNameSizeArray []NameSize

	//Проходим по всем файлам корневой директории.
	dirs, err := ioutil.ReadDir(rootPath)
	if err != nil {
		fmt.Println("Ошибка при чтении файлов ROOT директории:", err)
		return nameSizeArray, filesNameSizeArray, err
	}

	//Если очередной файл оказывается директорией, то запускается вычисление её размера
	//Размер и имя папки заносятся в ранее созданный nameSizeArray
	for _, dir := range dirs {
		if !dir.IsDir() {
			//Если файл не является директорией, то заносим его имя и размер в отдельный массив filesNameSizeArray
			file := dir
			fileSizeMb := float64(file.Size() / (1024 * 1024))
			fileNameSizeValue := NameSize{file.Name(), fileSizeMb}
			filesNameSizeArray = append(filesNameSizeArray, fileNameSizeValue)
		}
		//Вычисляем размер найденной директории
		c := make(chan float64) //Создаём канал, в который будут передаваться размеры найденных директорий
		defer close(c)
		go dirSizeCalculation(fmt.Sprintf("%s/%s", rootPath, dir.Name()), c)
		dirSize := <-c
		dirSizeMb := dirSize / (1024 * 1024)

		//Создаём переменную типа nameSize и добавления в nameSizeArray
		dirNameSizeValue := NameSize{dir.Name(), dirSizeMb}
		nameSizeArray = append(nameSizeArray, dirNameSizeValue)

		//Обработка возможной ошибки при возвращении в родительскую директорию
		err = os.Chdir("..")
		if err != nil {
			fmt.Println("..")
			return nameSizeArray, filesNameSizeArray, err
		}
	}

	return nameSizeArray, filesNameSizeArray, nil
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
