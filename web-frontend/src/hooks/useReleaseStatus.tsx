import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';

interface ReleaseStatusContextValue {
  released: boolean;
  isLoading: boolean;
  error: string | null;
}

const ReleaseStatusContext = createContext<ReleaseStatusContextValue | undefined>(undefined);

interface ReleaseStatusProviderProps {
  children: ReactNode;
}

export function ReleaseStatusProvider({ children }: ReleaseStatusProviderProps) {
  const [released, setReleased] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchReleaseStatus = async () => {
      try {
        const baseUrl = import.meta.env.VITE_API_BASE_URL || '';
        const response = await fetch(`${baseUrl}/api/v1/public/system/release-status`);
        if (!response.ok) {
          throw new Error('Failed to fetch release status');
        }
        const data = await response.json();
        setReleased(data.data.released);
      } catch (err) {
        console.error('Failed to fetch release status:', err);
        setError(err instanceof Error ? err.message : 'Unknown error');
        // Default to false (pre-release) on error
        setReleased(false);
      } finally {
        setIsLoading(false);
      }
    };

    fetchReleaseStatus();
  }, []);

  return (
    <ReleaseStatusContext.Provider value={{ released, isLoading, error }}>
      {children}
    </ReleaseStatusContext.Provider>
  );
}

export function useReleaseStatus(): ReleaseStatusContextValue {
  const context = useContext(ReleaseStatusContext);
  if (context === undefined) {
    // Return default values if used outside provider (for non-landing pages)
    return { released: true, isLoading: false, error: null };
  }
  return context;
}
