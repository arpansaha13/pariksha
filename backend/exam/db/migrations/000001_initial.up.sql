create table if not exists exams (
    id bigint primary key,
    title text not null,
    starts_at timestamp not null,
    ends_at timestamp not null,
    created_by bigint not null,
    type smallint default 1,
    max_candidates_count integer not null,
    max_score integer default 0,
    duration_minutes smallint not null check (duration_minutes >= 0 and duration_minutes <= 1440),
    paper_hash varchar(64),
    participant_counts jsonb default '{"unattended":0,"invited":0,"started":0,"ended":0}',
    hash varchar(64) not null unique,
    deleted_at timestamp
);

create table if not exists exam_permissions (
    exam_id bigint not null,
    user_id bigint not null,
    permissions smallint not null,
    primary key (exam_id, user_id),
    foreign key (exam_id) references exams(id) on delete cascade
);

create table if not exists exam_participants (
    id bigint primary key,
    exam_id bigint not null,
    user_id bigint not null,
    score_awarded integer default 0,
    status smallint default 1 not null,
    started_at timestamp,
    ended_at timestamp,
    scheduled_end_time timestamp,
    foreign key (exam_id) references exams(id) on delete cascade
);

create table if not exists exam_categories (
    id bigint primary key,
    exam_id bigint not null,
    category_id bigint not null,
    "order" smallint default 0 not null,
    foreign key (exam_id) references exams(id) on delete cascade
);

create table if not exists exam_questions (
    id bigint primary key,
    exam_id bigint not null,
    question_id bigint not null,
    category_id bigint not null,
    "order" smallint not null,
    max_score smallint not null,
    foreign key (exam_id) references exams(id) on delete cascade,
    unique(exam_id, question_id)
);

create table if not exists answers (
    id bigint primary key,
    exam_participant_id bigint not null,
    question_id bigint not null,
    answer jsonb,
    score_awarded smallint default 0 not null,
    evaluated boolean default false not null,
    foreign key (exam_participant_id) references exam_participants(id) on delete cascade
);
