package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/disintegration/letteravatar"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"image/jpeg"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func main() {
	ctx := context.Background()
	endpoint := "127.0.0.1:9000"
	accessKeyID := "masoud1"
	secretAccessKey := "Strong#Pass#2022"

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Предположим что к нам попадает либо фотография, которую хотят изменить, либо ничего в момент когда поменяли только ник
	// Проверка есть ли фотография, если есть изминить, если не то это изменили ник и по хорошему стоит изменить анонимную аватарку

	//var id uint32 = 3112

	// Создание базовых переменных
	name := "Hi"
	filename := "main.jpg"
	photo, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}

	rand.Seed(time.Now().UnixNano())

	// Создаём случайные запросы, что бы рандомизированно посмотреть на более приближенные к реальным условиям работы
	// 1 - создание нового пользователя
	// 2 - получение всех фотографий от конкретного пользователя
	// 3 - удаление пользователя
	var arr []int
	num := 10 // 10 запросов разных типов
	for k := 0; k < num; k++ {
		arr = append(arr, 1)
		arr = append(arr, 2)
		arr = append(arr, 3)
	}
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
	fmt.Println(arr)

	buckets, err := minioClient.ListBuckets(ctx)
	if err != nil {
		log.Fatal(err)
	}

	min_get_delete := 100
	max_get_delete := len(buckets)

	min_create := len(buckets) + 1
	max_create := min_create + 2000

	var id_make_user []int
	for len(id_make_user) < 10 {
		randomNum := rand.Intn(max_create-min_create+1) + min_create
		if !contains(id_make_user, randomNum) {
			id_make_user = append(id_make_user, randomNum)
		}
	}
	var id_get_user_photo []int
	for len(id_get_user_photo) < 10 {
		randomNum := rand.Intn(max_get_delete-min_get_delete+1) + min_get_delete
		if !contains(id_get_user_photo, randomNum) {
			id_get_user_photo = append(id_get_user_photo, randomNum)
		}
	}
	var id_delete_user []int
	for len(id_delete_user) < 10 {
		randomNum := rand.Intn(max_get_delete-min_get_delete+1) + min_get_delete
		if !contains(id_delete_user, randomNum) {
			id_delete_user = append(id_delete_user, randomNum)
		}
	}

	var time_out_create []int64
	var time_out_get []int64
	var time_out_delet []int64

	indx_create := 0
	indx_get := 0
	indx_delete := 0

	fmt.Println(arr)
	for _, v := range arr {
		switch v {
		case 1: // Создание нового пользователя
			start := time.Now()
			err = CreatingBasicMainUserPhoto_test(minioClient, ctx, uint32(id_make_user[indx_create]), name, photo)
			if err != nil {
				fmt.Println(err)
			}
			duration := time.Since(start).Milliseconds()
			time_out_create = append(time_out_create, duration)
			indx_create += 1

		case 2: // Получение фотографий
			start := time.Now()
			_, err = GetAllPhotoByIdMap_test(ctx, minioClient, uint32(id_get_user_photo[indx_get]))
			if err != nil {
				fmt.Println(err)
			}
			duration := time.Since(start).Milliseconds()
			time_out_get = append(time_out_get, duration)
			indx_get += 1
		case 3: // Удаление фотографий
			start := time.Now()
			err = DeleteAnyPhoto_test(minioClient, ctx, uint32(id_delete_user[indx_delete]), "main.jpg", name)
			if err != nil {
				fmt.Println(err)
			}
			duration := time.Since(start).Milliseconds()
			time_out_delet = append(time_out_delet, duration)
			indx_delete += 1
		}
	}
	fmt.Println("Время на создание пользователя", time_out_create)
	fmt.Println("Время на запросов к пользователю", time_out_get)
	fmt.Println("Время на удаление пользователя", time_out_delet)
}

func contains(nums []int, num int) bool {
	for _, n := range nums {
		if n == num {
			return true
		}
	}
	return false
}

func CreatingBasicMainUserPhoto_test(minioClient *minio.Client, ctx context.Context, id uint32, name string, photo []byte) error {
	bucket := strconv.Itoa(int(id))
	err := minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: "", ObjectLocking: true})
	if err != nil {
		return err
	}

	if photo != nil {
		photoReader := bytes.NewReader(photo)
		_, err := minioClient.PutObject(ctx, bucket, "main.jpg", photoReader, int64(photoReader.Len()), minio.PutObjectOptions{ContentType: "image/jpeg"})
		if err != nil {
			return err
		}
		return nil
	} else {
		firstLetter, _ := utf8.DecodeRuneInString(name)
		// Генерация аватарки по имени
		img, err := letteravatar.Draw(460, firstLetter, nil)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		err = jpeg.Encode(&buf, img, nil)
		if err != nil {
			return err
		}
		_, err = minioClient.PutObject(ctx, bucket, "main-anon.jpg", &buf, int64(buf.Len()), minio.PutObjectOptions{ContentType: "image/jpeg"})
		if err != nil {
			return err
		}
		return nil
	}
}

func renameFileInBucket_test(ctx context.Context, minioClient *minio.Client, bucket string, oldFileName string, newFileName string) error {
	srcOpts := minio.CopySrcOptions{
		Bucket: bucket,
		Object: oldFileName,
	}
	dstOpts := minio.CopyDestOptions{
		Bucket: bucket,
		Object: newFileName,
	}
	_, err := minioClient.CopyObject(ctx, dstOpts, srcOpts)
	if err != nil {
		return err
	}
	err = minioClient.RemoveObject(ctx, bucket, oldFileName, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func GetAllPhotoByIdMap_test(ctx context.Context, minioClient *minio.Client, id uint32) (map[string][]byte, error) {
	bucket := strconv.Itoa(int(id))
	objectCh := minioClient.ListObjects(ctx, bucket, minio.ListObjectsOptions{})

	var objects []minio.ObjectInfo
	map_photos := make(map[string][]byte)
	// Итерируемся по объектам и добавляем их в слайс
	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}
		objects = append(objects, object)
	}
	// Сортируем объекты по времени создания от новых к старым
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].LastModified.After(objects[j].LastModified)
	})
	// Выводим отсортированные объекты
	for _, object := range objects {
		img, err := minioClient.GetObject(ctx, bucket, object.Key, minio.GetObjectOptions{})
		if err != nil {
			return nil, err
		}
		photoBytes, err := io.ReadAll(img)
		if err != nil {
			return nil, err
		}
		map_photos[object.Key] = photoBytes
	}
	return map_photos, nil
}

// Удаление любой фотографии
func DeleteAnyPhoto_test(minioClient *minio.Client, ctx context.Context, id uint32, filename string, name string) error {
	bucket := strconv.Itoa(int(id))

	objectCh := minioClient.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    "img",
		Recursive: true,
	})

	var numbers []int
	for object := range objectCh {
		if object.Err != nil {
			return object.Err
		}
		numStr := object.Key[len("img") : len(object.Key)-len(".jpg")]
		num, err := strconv.Atoi(numStr)
		if err == nil {
			numbers = append(numbers, num)
		}
	}
	sort.Ints(numbers)
	nameW := strings.TrimSuffix(filename, filepath.Ext(filename))
	if nameW == "main" || nameW == "main-anon" {
		var maxNumber int
		if len(numbers) == 0 {
			// Этот сценарий подразумевает, то что замены удалённой фотографии нет и мы создаём анонимную на основе ника пользователя
			firstLetter, _ := utf8.DecodeRuneInString(name)
			// Генерация аватарки по имени
			img, err := letteravatar.Draw(460, firstLetter, nil)
			if err != nil {
				return err
			}
			var buf bytes.Buffer
			err = jpeg.Encode(&buf, img, nil)
			if err != nil {
				return err
			}
			_, err = minioClient.PutObject(ctx, bucket, "main-anon.jpg", &buf, int64(buf.Len()), minio.PutObjectOptions{ContentType: "image/jpeg"})
			if err != nil {
				return err
			}
			err = minioClient.RemoveObject(ctx, bucket, filename, minio.RemoveObjectOptions{})
			if err != nil {
				return err
			}
			return nil
		} else {
			// Мы бёрём максимально последнюю фотографию и переименовываем её в основную
			// TODO узнать на сколько критично, то что реализация сделана не через массив которы выстроили в порядке времени дополнения как это сделано в GetMap
			maxNumber = numbers[len(numbers)-1]
			maxMainFilename := "img" + strconv.Itoa(maxNumber) + ".jpg"
			err := renameFileInBucket_test(ctx, minioClient, bucket, maxMainFilename, filename)
			if err != nil {
				return err
			}
			return nil
		}
	} else {
		// Описано удаление обычной фотографии. Сначала проверка на существование, затем удаление
		bucket := strconv.Itoa(int(id))
		_, err := minioClient.StatObject(ctx, bucket, filename, minio.StatObjectOptions{})
		if err != nil {
			return err
		} else {
			opts := minio.RemoveObjectOptions{
				GovernanceBypass: true,
				VersionID:        "",
			}
			err = minioClient.RemoveObject(ctx, bucket, filename, opts)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func СhangeMainUserPhoto_test(minioClient *minio.Client, ctx context.Context, id uint32, name string, photo []byte) error {

	// TODO На сколько кричино важно что бы за рас удалялась не только одна фотография, то есть будет удалятся до момента минимального объёма фотографий
	bucket := strconv.Itoa(int(id))
	doneCh := make(chan struct{})
	defer close(doneCh)
	objectCh := minioClient.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    "img",
		Recursive: true,
	})

	var numbers []int
	for object := range objectCh {
		if object.Err != nil {
			return object.Err
		}
		numStr := object.Key[len("img") : len(object.Key)-len(".jpg")]
		num, err := strconv.Atoi(numStr)
		if err == nil {
			numbers = append(numbers, num)
		}
	}
	sort.Ints(numbers)
	// Проверка, на первое изменение main фотографии
	var maxNumber, minNumber int
	if len(numbers) == 0 {
		numbers = append(numbers, 1)
		maxNumber = 0
	} else {
		maxNumber = numbers[len(numbers)-1]
		minNumber = numbers[0]
	}

	//
	if photo != nil {
		_, err := minioClient.StatObject(ctx, bucket, "main.jpg", minio.StatObjectOptions{})
		if err != nil { // Создать main.jpg и удалить main-anon
			err = minioClient.RemoveObject(ctx, bucket, "main-anon.jpg", minio.RemoveObjectOptions{
				GovernanceBypass: true,
				VersionID:        "",
			})
			if err != nil {
				return err
			}
			photoReader := bytes.NewReader(photo)
			_, err := minioClient.PutObject(ctx, bucket, "main.jpg", photoReader, int64(photoReader.Len()), minio.PutObjectOptions{ContentType: "image/jpeg"})
			if err != nil {
				return err
			}
		} else {
			newMainPhoto := "img" + strconv.Itoa(maxNumber+1) + ".jpg"
			err = renameFileInBucket_test(ctx, minioClient, bucket, "main.jpg", newMainPhoto)
			if err != nil {
				return err
			}
			photoReader := bytes.NewReader(photo)
			_, err := minioClient.PutObject(ctx, bucket, "main.jpg", photoReader, int64(photoReader.Len()), minio.PutObjectOptions{ContentType: "image/jpeg"})
			if err != nil {
				return err
			}
		}
	} else {
		firstLetter, _ := utf8.DecodeRuneInString(name)
		// Генерация аватарки по имени
		img, err := letteravatar.Draw(460, firstLetter, nil)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		err = jpeg.Encode(&buf, img, nil)
		if err != nil {
			return err
		}
		_, err = minioClient.PutObject(ctx, bucket, "main-anon.jpg", &buf, int64(buf.Len()), minio.PutObjectOptions{ContentType: "image/jpeg"})
		if err != nil {
			return err
		}

	}
	// Проверка переполнение максимального количества фотографий в bucket, если больше 20 то удалить фото с минимальным индексом
	if len(numbers) <= 18 {
		return nil
	} else {
		// TODO записать цикл по удалению старх фотографий, то есть фотографий с маленьким индексом
		//Удаляем минимальную фотографию
		oldPhotoFilename := "img" + strconv.Itoa(minNumber) + ".jpg"
		err := minioClient.RemoveObject(ctx, bucket, oldPhotoFilename, minio.RemoveObjectOptions{
			GovernanceBypass: true,
			VersionID:        "",
		})
		if err != nil {
			return err
		}
	}
	return nil
}
