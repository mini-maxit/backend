CREATE TYPE registration_request_status AS ENUM ('approved', 'rejected', 'pending');

CREATE TYPE user_role AS ENUM ('student', 'teacher', 'admin');

CREATE TYPE submission_status AS ENUM ('received', 'sent for evaluation', 'evaluated', 'lost');
