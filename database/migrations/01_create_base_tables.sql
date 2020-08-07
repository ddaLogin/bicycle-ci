create table servers
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
    password text         not null,
    salt     text         not null,
    constraint users_login_uindex
        unique (login)
)
    comment 'Пользователи' charset = utf8;

create table images
(
    id          int auto_increment
        primary key,
    name        varchar(150) not null,
    description varchar(150) null,
    user_id     int          not null,
    constraint containers_users_id_fk
        foreign key (user_id) references users (id)
) charset = utf8;

create table providers
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
)
    comment 'VCS системы пользователей' charset = utf8;

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
    build_image     int          null,
    build_plan      text         null,
    artifact_dir    varchar(250) null,
    server_id       int          null,
    deploy_dir      varchar(250) null,
    constraint projects_name_uindex
        unique (name),
    constraint projects_providers_id_fk
        foreign key (provider) references vcs_providers (id),
    constraint projects_users_id_fk
        foreign key (user_id) references users (id)
) charset = utf8;

create table builds
(
    id         int auto_increment
        primary key,
    project_id int                                 not null,
    status     int       default 0                 not null,
    started_at timestamp default CURRENT_TIMESTAMP not null,
    ended_at   timestamp                           null,
    constraint builds_projects_id_fk
        foreign key (project_id) references projects (id)
) charset = utf8;

create table hooks
(
    id         int auto_increment
        primary key,
    project_id int          not null,
    user_id    int          not null,
    hook_id    varchar(150) null comment 'идентификатор хука на стороне провайдера',
    event      varchar(250) not null comment 'Событие при котором вызывается hook',
    branch     varchar(250) not null comment 'Ветка к которой привязан hook',
    constraint hooks_projects_id_fk
        foreign key (project_id) references projects (id),
    constraint hooks_users_id_fk
        foreign key (user_id) references users (id)
) charset = utf8;

create table steps
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