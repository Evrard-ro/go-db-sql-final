package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Failed to open DB") // настройте подключение к БД
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	lastID, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, lastID)

	// get
	gotParcel, err := store.Get(lastID)
	require.NoError(t, err)

	expectedParcel := parcel
	expectedParcel.Number = lastID

	assert.Equal(t, expectedParcel, gotParcel)
	// delete
	err = store.Delete(lastID)
	require.NoError(t, err)

	// try to get deleted parcel
	_, err = store.Get(lastID)
	require.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Failed to open DB") // настройте подключение к БД
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	lastId, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, lastId)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(lastId, newAddress)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	testParcel, err := store.Get(lastId)
	require.NoError(t, err)
	assert.Equal(t, newAddress, testParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Failed to open DB")
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	lastId, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, lastId)
	// set status
	// обновите статус, убедитесь в отсутствии ошибки

	newStatus := "delivered"
	err = store.SetStatus(lastId, newStatus)
	require.NoError(t, err)
	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	testParcel, err := store.Get(lastId)
	require.NoError(t, err)
	assert.Equal(t, newStatus, testParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Failed to open DB")
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		require.NoError(t, err)
		require.NotEmpty(t, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	require.NoError(t, err)
	require.Equal(t, len(parcels), len(storedParcels))
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных

	// check
	for _, parcel := range storedParcels {
		orig, ok := parcelMap[parcel.Number]
		require.True(t, ok, "Parcel with number %d not found in parcelMap", parcel.Number)
		assert.Equal(t, orig, parcel)
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
	}
}
