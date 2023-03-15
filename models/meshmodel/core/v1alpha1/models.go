package v1alpha1

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/layer5io/meshkit/database"
	"gorm.io/gorm"
)

var modelCreationLock sync.Mutex //Each component/relationship will perform a check and if the model already doesn't exist, it will create a model. This lock will make sure that there are no race conditions.
type ModelFilter struct {
	Name        string
	DisplayName string //If Name is already passed, avoid passing Display name unless greedy=true, else the filter will translate to an AND returning only the models where name and display name match exactly. Ignore, if this behaviour is expected.
	Greedy      bool   //when set to true - instead of an exact match, name will be prefix matched. Also an OR will be performed of name and display_name
	Version     string
	Category    string
	OrderOn     string
	Sort        string //asc or desc. Default behavior is asc
	Limit       int    //If 0 or  unspecified then all records are returned and limit is not used
	Offset      int
}

// swagger:response Model
type Model struct {
	ID          uuid.UUID              `json:"-"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	DisplayName string                 `json:"modelDisplayName" gorm:"modelDisplayName"`
	Category    Category               `json:"category"`
	SubCategory string                 `json:"subCategory" gorm:"subCategory"`
	Metadata    map[string]interface{} `json:"modelMetadata" yaml:"modelMetadata"`
}
type ModelDB struct {
	ID          uuid.UUID `json:"-"`
	CategoryID  uuid.UUID `json:"-" gorm:"modelID"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	DisplayName string    `json:"modelDisplayName" gorm:"modelDisplayName"`
	SubCategory string    `json:"subCategory" gorm:"subCategory"`
	Metadata    []byte    `json:"modelMetadata" gorm:"modelMetadata"`
}

// Create the filter from map[string]interface{}
func (cf *ModelFilter) Create(m map[string]interface{}) {
	if m == nil {
		return
	}
	cf.Name = m["name"].(string)
}

func CreateModel(db *database.Handler, cmodel Model) (uuid.UUID, error) {
	byt, err := json.Marshal(cmodel)
	if err != nil {
		return uuid.UUID{}, err
	}
	modelID := uuid.NewSHA1(uuid.UUID{}, byt)
	var model Model
	modelCreationLock.Lock()
	err = db.First(&model, "id = ?", modelID).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return uuid.UUID{}, err
	}
	if err == gorm.ErrRecordNotFound { //The model is already not present and needs to be inserted
		model = cmodel
		model.ID = modelID
		mdb := model.GetModelDB()
		err = db.Create(&mdb).Error
		if err != nil {
			modelCreationLock.Unlock()
			return uuid.UUID{}, err
		}
	}
	modelCreationLock.Unlock()
	return model.ID, nil
}
func (cmd *ModelDB) GetModel(cat Category) (c Model) {
	c.ID = cmd.ID
	c.Category = cat
	c.DisplayName = cmd.DisplayName
	c.Name = cmd.Name
	c.SubCategory = cmd.SubCategory
	c.Version = cmd.Version
	_ = json.Unmarshal(cmd.Metadata, &c.Metadata)
	return
}
func (c *Model) GetModelDB() (cmd ModelDB) {
	cmd.ID = c.ID
	cmd.DisplayName = c.DisplayName
	cmd.Name = c.Name
	cmd.SubCategory = c.SubCategory
	cmd.Version = c.Version
	cmd.Metadata, _ = json.Marshal(c.Metadata)
	return
}
