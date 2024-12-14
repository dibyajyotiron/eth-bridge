package repositories

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/eth-bridging/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	mockDb *sql.DB
	events = []*models.BridgeEvent{
		{
			TransactionHash: "0x3de4522433dd50f97857164bf36769b5196bc6856e9f5acc386f4ea1c199531d",
			Token:           "0x6B175474E89094C44Da98b954EedeAC495271d0F",
			Amount:          "88641847012511023937",
			FromChain:       "0xAa3a86C8Fc99b39F03dd0BCcc90f0F70DCC4cE17",
			ToChain:         "0xAa3a86C8Fc99b39F03dd0BCcc90f0F70DCC4cE17",
			Timestamp:       time.Now(),
		},
		{
			TransactionHash: "0x17d2cf21da2f3dbc56b1ae1278eeb864261cb487894f586a48a0590b14726558",
			Token:           "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE",
			Amount:          "32551712064712470",
			FromChain:       "0x2672a02DeA7A765545f4Bad9A7651c00EEa51ab2",
			ToChain:         "0x2672a02DeA7A765545f4Bad9A7651c00EEa51ab2",
			Timestamp:       time.Now(),
		},
	}
)

func TearDown(t *testing.T) {
	mockDb.Close()
}

func Setup(t *testing.T) (mock sqlmock.Sqlmock, repo BridgeEventRepository, gormDB *gorm.DB) {
	var err error
	mockDb, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	// defer mockDb.Close()

	dialector := postgres.New(postgres.Config{
		Conn:       mockDb,
		DriverName: "postgres",
	})

	// Create an instance of BridgeEventRepository with mocked DB
	gormDB, _ = gorm.Open(dialector, &gorm.Config{})

	repo = NewBridgeEventRepository(gormDB)
	return mock, repo, gormDB
}

func TestBridgeEventRepository_Save(t *testing.T) {
	// Create a mock database connection
	mock, repo, _ := Setup(t)
	defer TearDown(t)

	mock.ExpectBegin()
	// Mock the Create operation with RETURNING "id"
	mock.ExpectQuery(`INSERT INTO "bridge_events" (.+) VALUES (.+)`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
	mock.ExpectCommit()
	// Define the event to be saved

	// Call the Save method
	err := repo.Save(events[0])

	// Assert that no error occurred and the SQL statements were as expected
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBridgeEventRepository_GetAll(t *testing.T) {
	// Create a mock database connection
	mock, repo, _ := Setup(t)
	defer TearDown(t)

	// Mock the query
	rows := sqlmock.NewRows([]string{"id", "transaction_hash", "token", "amount", "from_chain", "to_chain", "timestamp"}).
		AddRow(events[0].ID, events[0].TransactionHash, events[0].Token, events[0].Amount, events[0].FromChain, events[0].ToChain, events[0].Timestamp).
		AddRow(events[1].ID, events[1].TransactionHash, events[1].Token, events[1].Amount, events[1].FromChain, events[1].ToChain, events[1].Timestamp)

	mock.ExpectQuery(`SELECT (.+) FROM "bridge_events" (.+)`).
		WillReturnRows(rows)

	// Call the GetAll method
	fetchedEvents, err := repo.GetAll(0, 2)

	// Assert that no error occurred, and the result matches the expected fetchedEvents
	assert.NoError(t, err)
	assert.Len(t, fetchedEvents, 2)
	assert.Equal(t, fetchedEvents, fetchedEvents)
	assert.NoError(t, mock.ExpectationsWereMet())
}
