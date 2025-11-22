create table if not exists categories (
    id bigint generated always as identity primary key,
    name varchar(255) not null,
    paper_indegree integer not null,
    exam_indegree integer not null,
    deleted_at timestamp
);

create table if not exists languages (
    id bigint generated always as identity primary key,
    slug varchar(255) not null unique,
    name varchar(255) not null,
    extension varchar(16) not null,
    version varchar(16) not null,
    is_enabled boolean not null default true
);

-- Seeding for languages
insert into languages (slug, name, extension, version)
values ('node', 'JavaScript (Node)', 'js', '22')
on conflict (slug) do nothing;

create table if not exists questions (
    id bigint generated always as identity primary key,
    question jsonb not null,
    type smallint not null check (type > 0 and type <= 3),
    hash varchar(64) not null unique,
    paper_indegree integer not null,
    exam_indegree integer not null,
    deleted_at timestamp
);

create table if not exists boilerplates (
    id bigint generated always as identity primary key,
    question_id bigint not null,
    language_id smallint not null,
    code text not null,
    foreign key (question_id) references questions(id) on delete set null,
    foreign key (language_id) references languages(id),
    unique (question_id, language_id)
);

create table if not exists test_cases (
    id bigint generated always as identity primary key,
    question_id bigint not null,
    "order" smallint not null,
    content jsonb not null,
    data_hash varchar(64) not null,
    hidden boolean not null default false,
    deleted_at timestamp,
    foreign key (question_id) references questions(id) on delete cascade,
    unique (question_id, "order")
);
