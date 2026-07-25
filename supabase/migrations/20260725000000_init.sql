-- Migration: Initialize profiles and attendance_logs tables

-- 1. Create Profiles Table (Users & Registered IP)
CREATE TABLE IF NOT EXISTS public.profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    registered_ip TEXT DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL
);

-- 2. Create Attendance Logs Table
CREATE TABLE IF NOT EXISTS public.attendance_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL,
    check_in_time TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
    latitude NUMERIC(10, 7) NOT NULL,
    longitude NUMERIC(10, 7) NOT NULL,
    distance_meters NUMERIC(10, 2) NOT NULL,
    ip_address TEXT NOT NULL,
    status TEXT NOT NULL,
    notes TEXT
);

-- Enable Row Level Security (RLS) or public access policies
ALTER TABLE public.profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.attendance_logs ENABLE ROW LEVEL SECURITY;

-- Create policies for full access (or adjust as needed)
CREATE POLICY "Allow all access to profiles" ON public.profiles FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "Allow all access to attendance_logs" ON public.attendance_logs FOR ALL USING (true) WITH CHECK (true);

-- 3. Insert Default 3 Users (chris, deksa, putra)
INSERT INTO public.profiles (username, password)
VALUES 
    ('chris', '12310299'),
    ('deksa', '12310300'),
    ('putra', '12310311')
ON CONFLICT (username) DO NOTHING;
