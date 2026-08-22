import { optimizeShoppingRoute, OptimizeResponse } from './api';

// Mock the global fetch function
globalThis.fetch = jest.fn();

describe('optimizeShoppingRoute', () => {
  beforeEach(() => {
    (globalThis.fetch as jest.Mock).mockClear();
  });

  it('successfully fetches and returns optimized route data', async () => {
    const mockResponseData: OptimizeResponse = {
      route_plan: {
        steps: [
          {
            shop_name: 'Aldi',
            items: ['tej'],
            coordinates: { latitude: 46.8451, longitude: 16.8455 }
          },
        ],
      },
      total_cost: 1500,
    };

    (globalThis.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponseData,
    });

    const result = await optimizeShoppingRoute('tej', 47.123, 19.456);

    expect(globalThis.fetch).toHaveBeenCalledWith(`${process.env.EXPO_PUBLIC_API_URL || 'https://localhost:8080'}/api/v1/optimize`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        user_input: 'tej',
        coords: {
          latitude: 47.123,
          longitude: 19.456,
        },
      }),
    });
    expect(result).toEqual(mockResponseData);
  });


  it('successfully uses EXPO_PUBLIC_API_URL environment variable', async () => {
    const originalEnv = process.env.EXPO_PUBLIC_API_URL;
    process.env.EXPO_PUBLIC_API_URL = 'https://api.example.com';

    const mockResponseData: OptimizeResponse = {
      route_plan: { steps: [] },
      total_cost: 0,
    };

    (globalThis.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponseData,
    });

    await optimizeShoppingRoute('tej', 47.123, 19.456);

    expect(globalThis.fetch).toHaveBeenCalledWith('https://api.example.com/api/v1/optimize', expect.any(Object));

    // Restore
    process.env.EXPO_PUBLIC_API_URL = originalEnv;
  });

  it('throws an error when the API returns a non-200 response with JSON error', async () => {
    (globalThis.fetch as jest.Mock).mockResolvedValueOnce({
      ok: false,
      status: 400,
      json: async () => ({ error: 'Invalid input', code: 400 }),
    });

    await expect(optimizeShoppingRoute('tej', 47.123, 19.456)).rejects.toThrow('Invalid input');
  });

  it('throws a generic error when the API returns a non-200 response without JSON', async () => {
    (globalThis.fetch as jest.Mock).mockResolvedValueOnce({
      ok: false,
      status: 500,
      json: async () => { throw new Error('Not JSON'); },
    });

    await expect(optimizeShoppingRoute('tej', 47.123, 19.456)).rejects.toThrow('HTTP error! status: 500');
  });
});
