package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func Test_parseOpenAPISpec(t *testing.T) {
	node, spec, err := parseOpenAPISpec("testdata/openapi.yaml")
	if err != nil {
		t.Errorf("parseOpenAPISpec() error = %v", err)
		return
	}
	assert.Equal(t, node.Kind, yaml.DocumentNode)
	assert.Equal(t, spec.Info.Title, "VRChat API Documentation")
}

func Test_getYAMLNodeKeys(t *testing.T) {
	_, _, err := parseOpenAPISpec("testdata/openapi.yaml")
	if err != nil {
		t.Errorf("parseOpenAPISpec() error = %v", err)
		return
	}
	type args struct {
		nodeKey string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "get keys from components",
			args: args{nodeKey: "components"},
			want: []string{"schemas", "securitySchemes", "responses", "parameters"},
		},
		{
			name: "get keys from components.schemas",
			args: args{nodeKey: "components.schemas"},
			want: strings.Split("UserExists Response Error AccountDeletionLog UserID BadgeID Badge AvatarID CurrentAvatarImageUrl CurrentAvatarThumbnailImageUrl Tag DeveloperType WorldID Platform PastDisplayName GroupID CurrentUserPresence UserState UserStatus CurrentUser TwoFactorAuthCode Verify2FAResult TwoFactorEmailCode Verify2FAEmailCodeResult VerifyAuthTokenResult Success ReleaseStatus UnityPackageID UnityPackage Avatar SortOption OrderOption CreateAvatarRequest UpdateAvatarRequest TransactionID TransactionStatus SubscriptionPeriod Subscription TransactionSteamWalletInfo TransactionSteamInfo TransactionAgreement Transaction LicenseGroupID UserSubscription LicenseType LicenseAction License LicenseGroup FavoriteID FavoriteType Favorite AddFavoriteRequest FavoriteGroupID FavoriteGroupVisibility FavoriteGroup UpdateFavoriteGroupRequest FileID MIMEType FileStatus FileData FileVersion File CreateFileRequest CreateFileVersionRequest FinishFileDataUploadRequest FileUploadURL FileVersionUploadStatus LimitedUser NotificationType Notification FriendStatus GroupShortCode GroupDiscriminator GroupMemberStatus GroupGalleryID GroupRoleID GroupGallery LimitedGroup GroupJoinState GroupPrivacy GroupRoleTemplate CreateGroupRequest GroupMemberID GroupMyMember GroupRole Group UpdateGroupRequest GroupAnnouncementID GroupAnnouncement CreateGroupAnnouncementRequest GroupAuditLogID GroupAuditLogEntry PaginatedGroupAuditLogEntryList GroupMemberLimitedUser GroupMember BanGroupMemberRequest CreateGroupGalleryRequest GroupGalleryImageID GroupGalleryImage UpdateGroupGalleryRequest AddGroupGalleryImageRequest InstanceID UdonProductId World GroupInstance CreateGroupInviteRequest GroupSearchSort GroupLimitedMember GroupUserVisibility UpdateGroupMemberRequest GroupRoleIDList GroupPermission NotificationID GroupPostVisibility GroupPost CreateGroupPostRequest GroupJoinRequestAction RespondGroupJoinRequest CreateGroupRoleRequest UpdateGroupRoleRequest InviteRequest SentNotification RequestInviteRequest InviteResponse InviteMessageType InviteMessageID InviteMessage UpdateInviteMessageRequest InstanceType InstanceRegion InstanceOwnerId GroupAccessType CreateInstanceRequest Region InstancePlatforms Instance InstanceShortNameResponse PermissionID Permission PlayerModerationID PlayerModerationType PlayerModeration ModerateUserRequest APIConfigAnnouncement DeploymentGroup APIConfigDownloadURLList DynamicContentRow APIConfigEvents APIConfig InfoPushDataClickable InfoPushDataArticleContent InfoPushDataArticle InfoPushData InfoPush APIHealth User UpdateUserRequest LimitedUserGroups representedGroup LimitedUnityPackage LimitedWorld CreateWorldRequest UpdateWorldRequest WorldMetadata WorldPublishStatus NotificationDetailInvite NotificationDetailInviteResponse NotificationDetailRequestInvite NotificationDetailRequestInviteResponse NotificationDetailVoteToKick", " "),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getYAMLNodeKeys(tt.args.nodeKey); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getYAMLNodeKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
