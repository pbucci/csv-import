package db

import (
	"github.com/guregu/null"
	"tableflow/go/pkg/model"
	"tableflow/go/pkg/tf"
)

func CreateObjectsForNewUser(user *model.User) (workspaceID, organizationID string, err error) {
	organization := &model.Organization{
		ID:        model.NewID(),
		Name:      "My Organization",
		CreatedBy: user.ID,
		UpdatedBy: user.ID,
	}
	organizationID = organization.ID.String()
	workspace := &model.Workspace{
		ID:             model.NewID(),
		OrganizationID: organization.ID,
		Name:           "My Workspace",
		CreatedBy:      user.ID,
		UpdatedBy:      user.ID,
	}
	workspaceID = workspace.ID.String()
	importer := &model.Importer{
		ID:          model.NewID(),
		WorkspaceID: workspace.ID,
		Name:        "Example Importer",
		CreatedBy:   user.ID,
		UpdatedBy:   user.ID,
		Template:    nil,
	}
	template := &model.Template{
		ID:          model.NewID(),
		WorkspaceID: workspace.ID,
		ImporterID:  importer.ID,
		Name:        "Default Template",
		CreatedBy:   user.ID,
		UpdatedBy:   user.ID,
	}
	templateColumns := &[]*model.TemplateColumn{
		{
			ID:                model.NewID(),
			TemplateID:        template.ID,
			Name:              "First Name",
			Key:               "first_name",
			Required:          false,
			DataType:          model.TemplateColumnDataTypeString,
			Description:       null.StringFrom("The user's first name"),
			SuggestedMappings: []string{"first"},
			CreatedBy:         user.ID,
			UpdatedBy:         user.ID,
		},
		{
			ID:                model.NewID(),
			TemplateID:        template.ID,
			Name:              "Last Name",
			Key:               "last_name",
			Required:          false,
			DataType:          model.TemplateColumnDataTypeString,
			Description:       null.StringFrom("The user's last name"),
			SuggestedMappings: []string{"last"},
			CreatedBy:         user.ID,
			UpdatedBy:         user.ID,
		},
		{
			ID:          model.NewID(),
			TemplateID:  template.ID,
			Name:        "Email",
			Key:         "email",
			Required:    true,
			DataType:    model.TemplateColumnDataTypeString,
			Description: null.StringFrom("The email of the user"),
			CreatedBy:   user.ID,
			UpdatedBy:   user.ID,
		},
	}
	err = tf.DB.Create(organization).Error
	if err != nil {
		tf.Log.Errorw("Error creating organization after sign up", "user_id", user.ID, "organization_id", organization.ID)
		return "", "", err
	}
	err = tf.DB.Create(workspace).Error
	if err != nil {
		tf.Log.Errorw("Error creating workspace after sign up", "user_id", user.ID, "workspace_id", workspace.ID)
		return "", "", err
	}
	err = tf.DB.Exec("insert into organization_users (organization_id, user_id) values (?, ?);", organization.ID, user.ID).Error
	if err != nil {
		tf.Log.Errorw("Error adding user to organization after sign up", "user_id", user.ID, "organization_id", organization.ID)
		return "", "", err
	}
	err = tf.DB.Exec("insert into workspace_users (workspace_id, user_id) values (?, ?);", workspace.ID, user.ID).Error
	if err != nil {
		tf.Log.Errorw("Error adding user to workspace after sign up", "user_id", user.ID, "workspace_id", workspace.ID)
		return "", "", err
	}
	err = tf.DB.Create(importer).Error
	if err != nil {
		tf.Log.Errorw("Error creating importer after sign up", "user_id", user.ID, "workspace_id", workspace.ID)
		return "", "", err
	}
	err = tf.DB.Create(template).Error
	if err != nil {
		tf.Log.Errorw("Error creating template after sign up", "user_id", user.ID, "workspace_id", workspace.ID)
		return "", "", err
	}
	err = tf.DB.Create(templateColumns).Error
	if err != nil {
		tf.Log.Errorw("Error creating template columns after sign up", "user_id", user.ID, "workspace_id", workspace.ID)
		return "", "", err
	}

	return workspaceID, organizationID, nil
}

func GetDatabaseSchemaInitSQL() string {
	return `
		create table if not exists instance_id (
		    id          uuid primary key not null,
		    initialized bool unique      not null default true,
		    constraint initialized check (initialized)
		);
		insert into instance_id (id) values (gen_random_uuid()) on conflict do nothing;

		create table if not exists organizations (
		    id         uuid primary key         not null default gen_random_uuid(),
		    name       text                     not null,
		    created_by uuid                     not null,
		    created_at timestamp with time zone not null,
		    updated_by uuid                     not null,
		    updated_at timestamp with time zone not null,
		    deleted_by uuid,
		    deleted_at timestamp with time zone
		);

		create table if not exists workspaces (
		    id              uuid primary key         not null default gen_random_uuid(),
		    organization_id uuid                     not null,
		    name            text                     not null,
		    api_key         text                     not null default concat('tf_', replace(gen_random_uuid()::text, '-', '')),
		    created_by      uuid                     not null,
		    created_at      timestamp with time zone not null,
		    updated_by      uuid                     not null,
		    updated_at      timestamp with time zone not null,
		    deleted_by      uuid,
		    deleted_at      timestamp with time zone,
		    constraint fk_organization_id
		        foreign key (organization_id)
		            references organizations(id)
		);
		create index if not exists workspaces_organization_id_idx on workspaces(organization_id);
		create unique index if not exists workspaces_api_key_idx on workspaces(api_key);

		create table if not exists organization_users (
		    organization_id uuid not null,
		    user_id         uuid not null,
		    primary key (organization_id, user_id),
		    constraint fk_organization_id
		        foreign key (organization_id)
		            references organizations(id)
		            on delete cascade
		            on update no action
		);
		create index if not exists organization_users_organization_id_idx on organization_users(organization_id);
		create index if not exists organization_users_user_id_idx on organization_users(user_id);

		create table if not exists workspace_users (
		    workspace_id uuid not null,
		    user_id      uuid not null,
		    primary key (workspace_id, user_id),
		    constraint fk_workspace_id
		        foreign key (workspace_id)
		            references workspaces(id)
		            on delete cascade
		            on update no action
		);
		create index if not exists workspace_users_workspace_id_idx on workspace_users(workspace_id);
		create index if not exists workspace_users_user_id_idx on workspace_users(user_id);

		create table if not exists importers (
		    id               uuid primary key         not null default gen_random_uuid(),
		    workspace_id     uuid                     not null,
		    name             text                     not null,
		    allowed_domains  text[]                   not null,
		    webhooks_enabled bool                     not null,
		    created_by       uuid                     not null,
		    created_at       timestamp with time zone not null,
		    updated_by       uuid                     not null,
		    updated_at       timestamp with time zone not null,
		    deleted_by       uuid,
		    deleted_at       timestamp with time zone,
		    constraint fk_workspace_id
		        foreign key (workspace_id)
		            references workspaces(id)
		);
		create index if not exists importers_workspace_id_created_at_idx on importers(workspace_id, created_at);

		create table if not exists templates (
		    id           uuid primary key         not null default gen_random_uuid(),
		    workspace_id uuid                     not null,
		    importer_id  uuid                     not null,
		    name         text                     not null,
		    created_by   uuid                     not null,
		    created_at   timestamp with time zone not null,
		    updated_by   uuid                     not null,
		    updated_at   timestamp with time zone not null,
		    deleted_by   uuid,
		    deleted_at   timestamp with time zone,
		    constraint fk_workspace_id
		        foreign key (workspace_id)
		            references workspaces(id),
		    constraint fk_importer_id_id
		        foreign key (importer_id)
		            references importers(id)
		);
		create index if not exists templates_workspace_id_created_at_idx on templates(workspace_id, created_at);
		create unique index if not exists templates_importer_id_idx on templates(importer_id) where (deleted_at is null);

		create table if not exists template_columns (
		    id          uuid primary key         not null default gen_random_uuid(),
		    template_id uuid                     not null,
		    name        text                     not null,
		    key         text                     not null,
		    required    bool                     not null default false,
		    index       int                      not null, -- The 0-based position of column
		    created_by  uuid                     not null,
		    created_at  timestamp with time zone not null,
		    updated_by  uuid                     not null,
		    updated_at  timestamp with time zone not null,
		    deleted_by  uuid,
		    deleted_at  timestamp with time zone,
		    constraint fk_template_id
		        foreign key (template_id)
		            references templates(id)
		);
		create unique index if not exists template_columns_template_id_key_idx on template_columns(template_id, key) where (deleted_at is null);
		create index if not exists template_columns_template_id_idx on template_columns(template_id);

		create table if not exists uploads (
		    id             uuid primary key not null default gen_random_uuid(),
		    tus_id         varchar(32)      not null,
		    importer_id    uuid             not null,
		    workspace_id   uuid             not null,
		    file_name      text,
		    file_type      text,
		    file_extension text,
		    file_size      bigint,
		    num_rows       int,
		    num_columns    int,
		    metadata       jsonb            not null default '{}'::jsonb, -- Optional custom data the client can send from the SDK, i.e. their user ID
		    is_parsed      bool             not null default false,       -- REMOVED: Are the upload_columns created and ready for the user to map?
		    is_stored      bool             not null default false,       -- Have all the records been stored?
		    error          text,
		    created_at     timestamptz      not null default now(),
            updated_at     timestamptz      not null default now(),
		    constraint fk_importer_id
		        foreign key (importer_id)
		            references importers(id),
		    constraint fk_workspace_id
		        foreign key (workspace_id)
		            references workspaces(id)
		);
		create unique index if not exists uploads_tus_id_idx on uploads(tus_id);
		create index if not exists uploads_workspace_id_created_at_idx on uploads(workspace_id, created_at);
		create index if not exists uploads_importer_id_idx on uploads(importer_id);

		create table if not exists upload_columns (
		    id                 uuid primary key not null default gen_random_uuid(),
		    upload_id          uuid             not null,
		    name               text             not null,
		    index              int              not null, -- The 0-based position of the column within the file
		    sample_data        text[]           not null,
		    template_column_id uuid,                      -- The ID of the template_column this column will map to when importing
		    unique (upload_id, index),
		    constraint fk_upload_id
		        foreign key (upload_id)
		            references uploads(id),
		    constraint fk_template_column_id              -- REMOVED: To support transient templates set on the upload
		        foreign key (template_column_id)
		            references template_columns(id)
		);
		create index if not exists upload_columns_upload_id_idx on upload_columns(upload_id);

		create table if not exists imports (
		    id                   uuid primary key not null default gen_random_uuid(),
		    upload_id            uuid             not null,
		    importer_id          uuid             not null,
		    workspace_id         uuid             not null,
		    file_type            text,
		    file_extension       text,
		    file_size            bigint,
		    num_rows             int,
		    num_columns          int,
		    num_processed_values int,
		    metadata             jsonb            not null default '{}'::jsonb, -- Optional custom data the client can send from the SDK, i.e. their user ID
		    is_stored            bool             not null default false,       -- Have all the records been stored?
		    created_at           timestamptz      not null default now(),
		    updated_at           timestamptz      not null default now(),
		    constraint fk_upload_id
		        foreign key (upload_id)
		            references uploads(id),
		    constraint fk_importer_id
		        foreign key (importer_id)
		            references importers(id),
		    constraint fk_workspace_id
		        foreign key (workspace_id)
		            references workspaces(id)
		);
		create index if not exists imports_workspace_id_created_at_idx on imports(workspace_id, created_at);
		create index if not exists imports_importer_id_idx on imports(importer_id);
		-- create unique index if not exists imports_upload_id_idx on imports(upload_id); -- REMOVED: To support imports being deleted

		create table if not exists validations (
		    id                 serial primary key,
		    template_column_id uuid  not null,
		    validate           text  not null,
		    options            jsonb,
		    message            text,
		    severity           text  not null default 'error',
		    deleted_at         timestamp with time zone,
		    constraint fk_template_column_id
		        foreign key (template_column_id)
		            references template_columns(id)
		);
		create index if not exists validations_template_column_id_idx on validations(template_column_id);


		/* Schema Update SQL */

		alter table uploads
		    drop column if exists storage_bucket;

		alter table imports
		    drop column if exists file_type;
		alter table imports
		    drop column if exists file_extension;
		alter table imports
		    drop column if exists file_size;
		alter table imports
		    drop column if exists storage_bucket;

		alter table importers
			add column if not exists webhooks_enabled bool not null default false;
		alter table importers
		    drop column if exists webhook_url;

		alter table template_columns
		    add column if not exists description text;

		alter table uploads
		    drop column if exists is_parsed;

		alter table uploads
		    add column if not exists header_row_index integer;
		alter table importers
		    add column if not exists skip_header_row_selection bool not null default false;

		alter table uploads
		    add column if not exists template jsonb;
		alter table upload_columns
		    drop constraint if exists fk_template_column_id;

		alter table uploads
		    add column if not exists schemaless bool not null default false;

		alter table imports
		    add column if not exists num_error_rows integer;
		alter table imports
		    add column if not exists num_valid_rows integer;

		alter table template_columns
		    add column if not exists suggested_mappings text[] not null default '{}';

		alter table imports
		    add column if not exists is_complete boolean default false not null;
		alter table imports
		    add column if not exists deleted_at timestamp with time zone;
		drop index if exists imports_upload_id_idx;
		create index if not exists imports_upload_id_non_unique_idx on imports(upload_id);
		create unique index if not exists imports_upload_id_deleted_at_idx on imports(upload_id) where (deleted_at is null);

		alter table template_columns
		    add column if not exists data_type text not null default 'string';
		alter table imports
		    add column if not exists data_types jsonb;
		do
		$$
		    begin
		        if exists (
		            select *
		            from information_schema.columns
		            where table_name = 'validations'
		              and column_name = 'type'
		        ) then
		            alter table validations
		                rename column type to validate;
		        end if;
		
		        if exists (
		            select *
		            from information_schema.columns
		            where table_name = 'validations'
		              and column_name = 'value'
		        ) then
		            alter table validations
		                rename column value to options;
		        end if;
		    end
		$$;
		alter table validations
		    alter column options drop not null;

		alter table uploads
		    add column if not exists sheet_list text[];

		do
		$$
			begin
				if not exists (select
				               from information_schema.columns
				               where table_name = 'uploads' and column_name = 'updated_at')
				then
					alter table uploads
						add column updated_at timestamptz;
		
					update uploads set updated_at = created_at where updated_at is null;
		
					alter table uploads
						alter column updated_at set default now();
					alter table uploads
						alter column updated_at set not null;
				end if;
			end
		$$;

		do
		$$
			begin
				if not exists (select
				               from information_schema.columns
				               where table_name = 'imports' and column_name = 'updated_at')
				then
					alter table imports
						add column updated_at timestamptz;
		
					update imports set updated_at = created_at where updated_at is null;
		
					alter table imports
						alter column updated_at set default now();
					alter table imports
						alter column updated_at set not null;
				end if;
			end
		$$;

		alter table uploads
		    add column if not exists matched_header_row_index integer;

		do
		$$
			declare
				column_exists boolean;
				r             record;
			begin
				select exists (select
				               from information_schema.columns
				               where table_name = 'template_columns' and column_name = 'index')
				into column_exists;
		
				-- If the index column does not exist, add it and perform the update
				if not column_exists
				then
					alter table template_columns
						add column index integer;
		
					for r in select template_id from template_columns where deleted_at is null group by template_id
						loop
							update template_columns
							set index = rn - 1
							from (select id, row_number() over (order by created_at) as rn
							      from template_columns
							      where template_id = r.template_id and deleted_at is null) as sub
							where template_columns.id = sub.id;
						end loop;
				end if;
			end
		$$;
	`
}
