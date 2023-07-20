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

func ArrayCreation(rootPath string) ([]NameSize, error) {
	//Срез структур с полями "имя" "размер" (для хранения имени и размера папок)
	var nameSizeArray []NameSize

	//Проходим по всем файлам указанной директории.
	dirs, err := ioutil.ReadDir(rootPath)
	if err != nil {
		fmt.Println("Ошибка при чтении файлов ROOT директории:", err)
		return nameSizeArray, err
	}

	//Если очередной файл явялется папкой, то вычисляем проводим необходимые действия
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		//Вычисляем размер папки
		c := make(chan float64)
		defer close(c)
		go dirSizeCalculation(fmt.Sprintf("%s/%s", rootPath, dir.Name()), c)
		dirSize := <-c
		dirSizeMb := dirSize / (1024 * 1024)
		//Создаём переменную типа nameSize и добавления в срез nameSizeArray (размер папки переводится в мегабайты!)
		nameSizeValue := NameSize{dir.Name(), dirSizeMb}
		nameSizeArray = append(nameSizeArray, nameSizeValue)

		//Обработка возможной ошибки при вовзращении в родительскую директорию
		err = os.Chdir("..")
		if err != nil {
			fmt.Println("..")
			return nameSizeArray, err
		}
	}

	return nameSizeArray, nil
}

func dirSizeCalculation(path string, c chan<- float64) {
	//Открываем канал sizes для передачи в него размеров вложенных дерикторий
	sizes := make(chan int64)

	//Данная функция считает размер всех файлов, которые не являются директориями, и отправляет результат канал sizes
	readSize := func(path string, file os.FileInfo, err error) error {
		if err != nil || file == nil {
			return nil
		}
		if !file.IsDir() {
			sizes <- file.Size()
		}
		return nil
	}

	//Каждая горутина считывает размер открытой для неё папки и отправляет результат в канал sizes
	//После завершения работы всег горутин канал закрывается
	go func() {
		filepath.Walk(path, readSize)
		close(sizes)
	}()

	//Суммируем всё, что находится в нашем канале sizes
	size := int64(0)
	for s := range sizes {
		size += s
	}

	//Возвращаем итоговый размер директории с учётом всех вложенных директорий
	c <- float64(size)
}
