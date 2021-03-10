package api

import (
	"bytes"
	"eco/services/halt"
	"eco/services/session"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

func getCentrosDeCosto(s *session.EcoSession) error {

	w, _ := s.GetHttp()

	/*type CentroCosto struct {
		Nombre                  string `"json:"nombre"`
		Codigo                  string `"json:"codigo"`
		Descripcion             string `"json:"descripcion"`
		Activo                  bool   `"json:"activo"`
		EsCentroBeneficio       bool   `"json:"esCentroBeneficio"`
		PlanificacionObraItemId int    `"json:"planificacionObraItemId"`
	}

	var centroCosto CentroCosto*/

	//ecoApi := net.NewEcoApi(s, "teamplace", "/api/1/teamplace/CentroCostoCons")

	/*params := make(map[string]string)
	params["diccAlias"] = "CLIENTE"
	ecoApi.QueryParameters(params)*/

	/*if err := ecoApi.Get("", &centroCosto); err != nil {
		return err
	}*/

	type User struct {
		Id       int    `"json:"id"`
		Name     string `"json:"name"`
		Username string `"json:"username"`
		Email    string `"json:"email"`
	}

	var users []User

	endpoint := "https://jsonplaceholder.typicode.com/users"
	method := "GET"

	res, err := httpRequest(endpoint, method, nil, w)
	if err != nil {
		halt.Error(err, http.StatusInternalServerError).HTTPError(w)
	}

	if res.StatusCode != 200 {
		halt.Error(err, http.StatusInternalServerError).HTTPError(w)
	}

	defer res.Body.Close()

	json.NewDecoder(res.Body).Decode(&users)

	json.NewEncoder(w).Encode(&users)

	return nil
}

func getCentroDeCosto(s *session.EcoSession) error {

	w, r := s.GetHttp()

	vars := mux.Vars(r)
	userId := vars["id"]

	type User struct {
		Id       int    `"json:"id"`
		Name     string `"json:"name"`
		Username string `"json:"username"`
		Email    string `"json:"email"`
	}

	var user User

	endpoint := "https://jsonplaceholder.typicode.com/users/" + userId
	method := "GET"

	res, err := httpRequest(endpoint, method, nil, w)
	if err != nil {
		halt.Error(err, http.StatusInternalServerError).HTTPError(w)
	}

	if res.StatusCode != 200 {
		halt.Error(err, http.StatusInternalServerError).HTTPError(w)
	}

	defer res.Body.Close()

	json.NewDecoder(res.Body).Decode(&user)

	json.NewEncoder(w).Encode(&user)

	return nil
}

func httpRequest(endpoint string, method string, input interface{}, w http.ResponseWriter) (*http.Response, error) {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(&input)

	req, err := http.NewRequest(method, endpoint, buf)
	if err != nil {
		halt.Error(err, http.StatusInternalServerError).HTTPError(w)
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	res, err := client.Do(req)

	return res, err
}

/*func getEndpoint(apiUrl string) string {
	urlPrefix, _ := config.EcoCfg.String("environment.host")
	endpoint := urlPrefix + apiUrl

	return endpoint
}*/

/*func search(s *session.EcoSession) error {

	w, r := s.GetHttp()

	var searchedData = r.URL.Query().Get("query")
	var response models.SearchData
	var tokenInfo models.InfoToken
	var tree interface{}
	var err error

	ecoApiInfoToken := net.NewEcoApi(s, "oauth", "/auth/token/info")
	if err := ecoApiInfoToken.Get("", &tokenInfo); err != nil {
		return err
	}

	if tokenInfo.Admin && tokenInfo.Internal {

		tree, err = getTree(s)
		if err == nil {
			var teamplaceMenuOptions []models.MenuOption
			treeTraversal(tree, &teamplaceMenuOptions)
			response.TeamplaceMenuOptions = filterMenuOptions(teamplaceMenuOptions, searchedData)
		}

		response.Clients, err = filterClients(s, searchedData)
		if err != nil {
			return err
		}

		response.Providers, err = filterProviders(s, searchedData)
		if err != nil {
			return err
		}

		response.Users, err = filterUsers(s, searchedData)
		if err != nil {
			return err
		}

	}

	response.Features, err = filterFeatures(s, searchedData)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(&response)
}

func getTree(s *session.EcoSession) (interface{}, error) {

	endPoint := net.NewEcoApi(s, "teamplace", "/api/1/teamplace/arbol/info")

	var tree interface{}
	if err := endPoint.Get("", &tree); err != nil {
		return nil, err
	}

	return tree, nil
}

func treeTraversal(data interface{}, teamplaceMenuOptions *[]models.MenuOption) {

	rData := reflect.ValueOf(data)

	if reflect.TypeOf(data).Kind() == reflect.Slice {
		for i := 0; i < rData.Len(); i++ {
			treeTraversal(rData.Index(i).Interface(), teamplaceMenuOptions)
		}
		return
	}

	if reflect.TypeOf(data).Kind() == reflect.Map {
		for _, value := range rData.MapKeys() {

			if value.String() == "arbol" || value.String() == "nodo" {
				node := rData.MapIndex(value)
				treeTraversal(node.Interface(), teamplaceMenuOptions)
			}

			// Caso base
			if value.String() == "hoja" {

				leaf := rData.MapIndex(value)
				rLeaf := reflect.ValueOf(leaf.Interface())

				if reflect.TypeOf(leaf.Interface()).Kind() == reflect.Slice {
					for i := 0; i < rLeaf.Len(); i++ {
						rLeafValue := reflect.ValueOf(rLeaf.Index(i).Interface())
						menuOption := getMenuOption(rLeafValue)
						*teamplaceMenuOptions = append(*teamplaceMenuOptions, menuOption)
					}
				}

				if reflect.TypeOf(leaf.Interface()).Kind() == reflect.Map {
					rLeafValue := reflect.ValueOf(rLeaf.Interface())
					menuOption := getMenuOption(rLeafValue)
					*teamplaceMenuOptions = append(*teamplaceMenuOptions, menuOption)
				}

			}

		}
	}
}

func getMenuOption(rLeafValue reflect.Value) models.MenuOption {
	var leafData models.MenuOptionlk
	for _, value := range rLeafValue.MapKeys() {

		switch value.String() {
		case "id":
			idStringValue := fmt.Sprintf("%f", rLeafValue.MapIndex(value))
			idFloatValue, _ := strconv.ParseFloat(idStringValue, 32)
			idIntValue := int(idFloatValue)
			leafData.Id = idIntValue
		case "tipo":
			typeStringValue := fmt.Sprintf("%f", rLeafValue.MapIndex(value))
			typeFloatValue, _ := strconv.ParseFloat(typeStringValue, 32)
			typeIntValue := int(typeFloatValue)
			leafData.Type = typeIntValue
		case "caption":
			caption := fmt.Sprintf("%s", rLeafValue.MapIndex(value))
			leafData.Caption = caption
		case "shortCaption":
			shortCaption := fmt.Sprintf("%s", rLeafValue.MapIndex(value))
			leafData.ShortCaption = shortCaption
		}

	}

	return leafData
}

func filterMenuOptions(teamplaceMenuOptions []models.MenuOption, searchedData string) []models.MenuOption {

	var filteredMenuOptions []models.MenuOption

	for _, menuOption := range teamplaceMenuOptions {
		if containsSearchedData(menuOption.ShortCaption, searchedData) {
			filteredMenuOptions = append(filteredMenuOptions, menuOption)
		}
	}

	return filteredMenuOptions
}

func filterFeatures(s *session.EcoSession, searchedData string) ([]models.FeatureData, error) {

	var contextList []models.ContextList
	var arrayFeatureData []models.FeatureData

	ecoApi := net.NewEcoApi(s, "contexts", "/api/1/contexts")

	if err := ecoApi.Get("", &contextList); err != nil {
		return nil, err
	}

	for _, context := range contextList {

		var contextData persistence.ContextData
		var featureData models.FeatureData

		ecoApi := net.NewEcoApi(s, "contexts", "/api/1/contexts/")
		if err := ecoApi.Get(context.ID, &contextData); err != nil {
			return nil, err
		}

		for _, feature := range contextData.Features {

			if containsSearchedData(feature.Name, searchedData) {
				featureData.FeatureId = feature.ID
				featureData.FeatureName = feature.Name
				featureData.FeatureKind = feature.Kind
				featureData.ContextId = contextData.ContextBase.ID
				featureData.ContextName = contextData.ContextBase.Name
				featureData.ContextColor = contextData.Color
				arrayFeatureData = append(arrayFeatureData, featureData)
			}

		}

	}

	return arrayFeatureData, nil
}

func filterClients(s *session.EcoSession, searchedData string) ([]models.ClientData, error) {

	var clientList []models.ClientData
	var arrayClientData []models.ClientData

	ecoApi := net.NewEcoApi(s, "teamplace", "/api/1/teamplace/filters")

	params := make(map[string]string)
	params["diccAlias"] = "CLIENTE"
	ecoApi.QueryParameters(params)

	if err := ecoApi.Get("", &clientList); err != nil {
		return nil, err
	}

	for _, client := range clientList {
		if containsSearchedData(client.Caption, searchedData) {
			arrayClientData = append(arrayClientData, client)
		}
	}

	return arrayClientData, nil
}

func filterProviders(s *session.EcoSession, searchedData string) ([]models.ProviderData, error) {

	var providerList []models.ProviderData
	var arrayProviderData []models.ProviderData

	ecoApi := net.NewEcoApi(s, "teamplace", "/api/1/teamplace/filters")

	params := make(map[string]string)
	params["diccAlias"] = "PROVEEDOR"
	ecoApi.QueryParameters(params)

	if err := ecoApi.Get("", &providerList); err != nil {
		return nil, err
	}

	for _, provider := range providerList {
		if containsSearchedData(provider.Caption, searchedData) {
			arrayProviderData = append(arrayProviderData, provider)
		}
	}

	return arrayProviderData, nil
}

func filterUsers(s *session.EcoSession, searchedData string) ([]models.UserData, error) {

	var userList []models.UserData
	var arrayUserData []models.UserData

	apiEndpoint := fmt.Sprintf("/api/1/users/%s", s.Domain)
	ecoApi := net.NewEcoApi(s, "users", apiEndpoint)

	if err := ecoApi.Get("", &userList); err != nil {
		return nil, err
	}

	for _, user := range userList {
		userFullName := user.FirstName + " " + user.LastName
		if containsSearchedData(userFullName, searchedData) {
			arrayUserData = append(arrayUserData, user)
		}
	}

	return arrayUserData, nil
}

func containsSearchedData(value string, filter string) bool {
	return strings.Contains(strings.ToLower(removeAccents(value)), strings.ToLower(removeAccents(filter)))
}

func removeAccents(value string) string {
	transformer := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(transformer, value)

	return result
}
*/
