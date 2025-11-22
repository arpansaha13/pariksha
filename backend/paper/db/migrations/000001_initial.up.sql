create table if not exists papers (
    id bigint generated always as identity primary key,
    title varchar(255) not null default 'Untitled Paper',
    duration_minutes smallint not null check (duration_minutes >= 0 and duration_minutes <= 1440),
    created_by bigint not null,
    hash varchar(64) not null unique,
    deleted_at timestamp
);

create table if not exists permissions (
    paper_id bigint not null,
    user_id bigint not null,
    permissions smallint not null,
    deleted_at timestamp,
    primary key (paper_id, user_id),
    foreign key (paper_id) references papers(id) on delete cascade
);

create table if not exists paper_categories (
    paper_id bigint not null,
    category_id bigint not null,
    "order" smallint not null,
    primary key (paper_id, category_id),
    foreign key (paper_id) references papers(id) on delete cascade
);

create table if not exists paper_questions (
    paper_id bigint not null,
    question_id bigint not null,
    category_id bigint not null,
    "order" smallint not null,
    max_score smallint not null check (max_score >= 0 and max_score <= 1000),
    primary key (paper_id, question_id),
    foreign key (paper_id) references papers(id) on delete cascade
);
