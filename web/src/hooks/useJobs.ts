import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../services/api';

export interface JobStats {
    total_jobs: number;
    pending_jobs: number;
    running_jobs: number;
    failed_jobs: number;
    completed_jobs: number;
}

export interface Job {
    id: string;
    type: string;
    status: 'pending' | 'running' | 'completed' | 'failed';
    created_at: string;
    retry_count: number;
    payload?: string;
    error_message?: string;
}

export interface JobsResponse {
    jobs: Job[];
    meta: {
        current_page: number;
        total_pages: number;
        total_records: number;
    };
}

export interface CreateJobPayload {
    type: string;
    payload: string;
}

// Hooks

interface RawJobData {
    jobId?: string;
    job_id?: string;
    type: string;
    status: 'pending' | 'running' | 'completed' | 'failed';
    createdAt?: string;
    created_at?: string;
    retryCount?: string;
    retry_count?: string;
    payload?: string;
    errorMessage?: string;
    error_message?: string;
}

interface RawMetaData {
    currentPage?: number;
    current_page?: number;
    totalPages?: number;
    total_pages?: number;
    totalRecords?: string | number;
    total_records?: string | number;
}

interface RawResponse {
    jobs: RawJobData[];
    meta: RawMetaData;
}

interface RawJobStats {
    totalJobs?: string;
    total_jobs?: string;
    pendingJobs?: string;
    pending_jobs?: string;
    runningJobs?: string;
    running_jobs?: string;
    failedJobs?: string;
    failed_jobs?: string;
    completedJobs?: string;
    completed_jobs?: string;
}

const mapJob = (raw: RawJobData): Job => ({
    id: raw.jobId || raw.job_id || '',
    type: raw.type,
    status: raw.status,
    created_at: raw.createdAt || raw.created_at || '',
    retry_count: parseInt((raw.retryCount || raw.retry_count || '0').toString(), 10),
    payload: raw.payload,
    error_message: raw.errorMessage || raw.error_message,
});

export const useJobStats = () => {
    return useQuery({
        queryKey: ['jobStats'],
        queryFn: async () => {
            const { data } = await api.get<RawJobStats>('/stats');
            return {
                total_jobs: parseInt((data.totalJobs ?? data.total_jobs ?? '0').toString(), 10),
                pending_jobs: parseInt((data.pendingJobs ?? data.pending_jobs ?? '0').toString(), 10),
                running_jobs: parseInt((data.runningJobs ?? data.running_jobs ?? '0').toString(), 10),
                failed_jobs: parseInt((data.failedJobs ?? data.failed_jobs ?? '0').toString(), 10),
                completed_jobs: parseInt((data.completedJobs ?? data.completed_jobs ?? '0').toString(), 10),
            };
        },
        refetchInterval: 2000,
    });
};

export const useJobs = (limit: number = 20, offset: number = 0) => {
    return useQuery({
        queryKey: ['jobs', limit, offset],
        queryFn: async () => {
            const { data } = await api.get<RawResponse>(`/jobs`, {
                params: { limit, offset },
            });
            return {
                jobs: (data.jobs || []).map(mapJob),
                meta: {
                    current_page: data.meta.currentPage ?? data.meta.current_page ?? 1,
                    total_pages: data.meta.totalPages ?? data.meta.total_pages ?? 1,
                    total_records: parseInt((data.meta.totalRecords ?? data.meta.total_records ?? '0').toString(), 10),
                },
            };
        },
        refetchInterval: 2000,
    });
};

export const useDeadJobs = (limit: number = 20, offset: number = 0) => {
    return useQuery({
        queryKey: ['deadJobs', limit, offset],
        queryFn: async () => {
            const { data } = await api.get<RawResponse>(`/jobs/dead`, {
                params: { limit, offset },
            });
            return {
                jobs: (data.jobs || []).map(mapJob),
                meta: {
                    current_page: data.meta.currentPage ?? data.meta.current_page ?? 1,
                    total_pages: data.meta.totalPages ?? data.meta.total_pages ?? 1,
                    total_records: parseInt((data.meta.totalRecords ?? data.meta.total_records ?? '0').toString(), 10),
                },
            };
        },
        refetchInterval: 2000,
    });
};

export const useJob = (id: string) => {
    return useQuery({
        queryKey: ['job', id],
        queryFn: async () => {
            const { data } = await api.get<RawJobData>(`/jobs/${id}`);
            return mapJob(data);
        },
        enabled: !!id,
    });
};

export const useCreateJob = () => {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: async (newJob: CreateJobPayload) => {
            const { data } = await api.post('/jobs', newJob);
            return data;
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['jobs'] });
            queryClient.invalidateQueries({ queryKey: ['jobStats'] });
        },
    });
};
