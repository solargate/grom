package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solargate/travka/internal/social"
)

type UserSearchResult struct {
	Nickname string `json:"nickname" example:"bob"`
	Name     string `json:"name" example:"Bob"`
	Handle   string `json:"handle" example:"bob@travka.example"`
	IsLocal  bool   `json:"is_local" example:"true"`
}

func toUserSearchResults(items []social.UserSearchResult) []UserSearchResult {
	result := make([]UserSearchResult, len(items))
	for i := range items {
		result[i] = UserSearchResult{
			Nickname: items[i].Nickname,
			Name:     items[i].Name,
			Handle:   items[i].Handle,
			IsLocal:  items[i].IsLocal,
		}
	}
	return result
}

// searchUsers godoc
// @Summary      Search users
// @Description  Search local users by nickname or resolve a federated handle
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        q  query  string  true  "Search query"
// @Success      200  {array}  UserSearchResult
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /users/search [get]
func searchUsers(ctx *gin.Context) {
	if err := initSocialService(); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to init social service"})
		return
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	query := ctx.Query("q")
	results, err := socialService.SearchLocal(query, userID)
	if err != nil {
		handleSocialError(ctx, err)
		return
	}
	if results == nil {
		ctx.JSON(http.StatusOK, []UserSearchResult{})
		return
	}
	ctx.JSON(http.StatusOK, toUserSearchResults(results))
}
