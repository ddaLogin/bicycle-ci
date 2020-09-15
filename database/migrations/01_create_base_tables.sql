create table remote_servers
(
    id             int auto_increment
        primary key,
    name           varchar(250) not null,
    login          varchar(150) not null,
    host           varchar(150) not null,
    deploy_public  text         not null,
    deploy_private text         not null
) charset = utf8;

create table users
(
    id       int auto_increment
        primary key,
    login    varchar(255) not null,
    password mediumtext   not null,
    salt     mediumtext   not null,
    constraint users_login_uindex
        unique (login)
) comment 'Пользователи' charset = utf8;;

create table docker_images
(
    id          int auto_increment
        primary key,
    name        varchar(150) not null,
    description varchar(150) null,
    user_id     int          not null,
    constraint containers_users_id_fk
        foreign key (user_id) references users (id)
) charset = utf8;

create table vcs_providers
(
    id                     int auto_increment
        primary key,
    user_id                int          not null,
    provider_type          int          not null,
    provider_auth_token    text         not null,
    provider_account_id    int          not null,
    provider_account_login varchar(255) not null,
    constraint providers_user_id_provider_type_uindex
        unique (user_id, provider_type),
    constraint providers_users_id_fk
        foreign key (user_id) references users (id)
) comment 'VCS системы пользователей' charset = utf8;

create table projects
(
    id              int auto_increment
        primary key,
    user_id         int          not null,
    name            varchar(255) not null,
    provider        int          not null,
    repo_id         int          not null,
    repo_name       varchar(255) not null,
    repo_owner_name varchar(255) not null,
    repo_owner_id   varchar(255) not null,
    deploy_key_id   int          null,
    deploy_private  text         null,
    constraint projects_name_uindex
        unique (name),
    constraint projects_providers_id_fk
        foreign key (provider) references vcs_providers (id),
    constraint projects_users_id_fk
        foreign key (user_id) references users (id)
) charset = utf8;

create table project_build_plans
(
    id                int auto_increment
        primary key,
    project_id        int            not null,
    title             varbinary(255) not null,
    docker_image_id   int            not null,
    build_instruction text           not null,
    artifact          varchar(255)   not null,
    constraint project_build_plans_project_id_title_uindex
        unique (project_id, title),
    constraint project_build_plans_images_id_fk
        foreign key (docker_image_id) references docker_images (id),
    constraint project_build_plans_projects_id_fk
        foreign key (project_id) references projects (id)
) charset = utf8;

create table builds
(
    id                    int auto_increment
        primary key,
    project_build_plan_id int                                 not null,
    user_id               int                                 not null,
    branch                varchar(60)                         not null,
    status                int       default 0                 not null,
    started_at            timestamp default CURRENT_TIMESTAMP not null,
    ended_at              timestamp                           null,
    constraint builds_project_build_plans_id_fk
        foreign key (project_build_plan_id) references project_build_plans (id),
    constraint builds_users_id_fk
        foreign key (user_id) references users (id)
) charset = utf8;

create table build_steps
(
    id       int auto_increment
        primary key,
    build_id int          not null,
    name     varchar(250) not null,
    std_out  text         null,
    std_err  text         null,
    error    text         null,
    status   int          not null,
    constraint steps_builds_id_fk
        foreign key (build_id) references builds (id)
) charset = utf8;

create table project_deploy_plans
(
    id                   int auto_increment
        primary key,
    project_id           int          not null,
    title                varchar(255) not null,
    remote_server_id     int          null,
    deployment_directory varchar(255) not null,
    constraint project_deploy_plans_project_id_title_uindex
        unique (project_id, title),
    constraint project_deploy_plans_projects_id_fk
        foreign key (project_id) references projects (id),
    constraint project_deploy_plans_servers_id_fk
        foreign key (remote_server_id) references remote_servers (id)
) charset = utf8;

create table vcs_hooks
(
    id                    int auto_increment
        primary key,
    project_id            int          not null,
    project_build_plan_id int          not null,
    user_id               int          not null,
    hook_id               varchar(150) not null comment 'идентификатор хука на стороне провайдера',
    event                 varchar(250) not null comment 'Событие при котором вызывается hook',
    branch                varchar(250) not null comment 'Ветка к которой привязан hook',
    constraint hooks_projects_id_fk
        foreign key (project_id) references projects (id),
    constraint hooks_users_id_fk
        foreign key (user_id) references users (id),
    constraint vcs_hooks_project_build_plans_id_fk
        foreign key (project_build_plan_id) references project_build_plans (id)
) charset = utf8;

create table deployments
(
    id                     int auto_increment
        primary key,
    project_deploy_plan_id int                                 not null,
    build_id               int                                 not null,
    user_id                int                                 not null,
    status                 int       default 0                 not null,
    std_out                text                                null,
    std_err                text                                null,
    error                  text                                null,
    started_at             timestamp default CURRENT_TIMESTAMP not null,
    ended_at               timestamp                           null,
    constraint deployments_builds_id_fk
        foreign key (build_id) references builds (id),
    constraint deployments_project_deploy_plans_id_fk
        foreign key (project_deploy_plan_id) references project_deploy_plans (id),
    constraint deployments_users_id_fk
        foreign key (user_id) references users (id)
)  charset = utf8;

