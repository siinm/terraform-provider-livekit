// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/livekit/protocol/auth"
)

var _ resource.Resource = &AccessTokenResource{}
var _ resource.ResourceWithImportState = &AccessTokenResource{}

func NewAccessTokenResource() resource.Resource {
	return &AccessTokenResource{}
}

// AccessTokenResource defines the resource implementation.
type AccessTokenResource struct {
	apiKeys *auth.AccessToken
}

// AccessTokenResourceModel describes the resource data model.
type AccessTokenResourceModel struct {
	Room           types.String `tfsdk:"room"`
	Identity       types.String `tfsdk:"identity"`
	CanPublish     types.Bool   `tfsdk:"can_publish"`
	CanPublishData types.Bool   `tfsdk:"can_publish_data"`
	CanSubscribe   types.Bool   `tfsdk:"can_subscribe"`
	ValidFor       types.String `tfsdk:"valid_for"`
	Token          types.String `tfsdk:"token"`
}

func (r *AccessTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_token"
}

func (r *AccessTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Access Token",

		Attributes: map[string]schema.Attribute{
			"room": schema.StringAttribute{
				MarkdownDescription: "Room name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"identity": schema.StringAttribute{
				MarkdownDescription: "Token identity to connect into the room",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"can_publish": schema.BoolAttribute{
				MarkdownDescription: "Can Publish",
				Required:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"can_publish_data": schema.BoolAttribute{
				MarkdownDescription: "Can publish data",
				Required:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"can_subscribe": schema.BoolAttribute{
				MarkdownDescription: "Can subscribe",
				Required:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"valid_for": schema.StringAttribute{
				MarkdownDescription: "Validity duration of the token, e.g. 1h, 1d, 1w, etc.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("1h"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The generated JWT token",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *AccessTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	apiKeys, ok := req.ProviderData.(*auth.AccessToken)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *auth.AccessToken, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.apiKeys = apiKeys
}

func (r *AccessTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessTokenResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	validFor, err := time.ParseDuration(data.ValidFor.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid valid_for ", err.Error())
		return
	}

	grant := &auth.VideoGrant{
		Room:           data.Room.ValueString(),
		CanPublish:     data.CanPublish.ValueBoolPointer(),
		CanPublishData: data.CanPublishData.ValueBoolPointer(),
		CanSubscribe:   data.CanSubscribe.ValueBoolPointer(),
		RoomJoin:       true,
	}

	at := r.apiKeys.AddGrant(grant).
		SetIdentity(data.Identity.ValueString()).
		SetValidFor(validFor)

	jwt, err := at.ToJWT()
	if err != nil {
		resp.Diagnostics.AddError("Error creating JWT", err.Error())
		return
	}

	data.Token = types.StringValue(jwt)

	tflog.Trace(ctx, "created a resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccessTokenResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccessTokenResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// nothing to do, always requires replacement when field changes.

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessTokenResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// the livekit API does not support deleting tokens, so we don't need to do anything here
}

func (r *AccessTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
