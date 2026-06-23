import { optimizeShoppingRoute, OptimizeResponse } from './api';

// Mock the global fetch function
global.fetch = jest.fn();

describe('optimizeShoppingRoute', () => {
  beforeEach(() => {
    (global.fetch as jest.Mock).mockClear();
  });

  it('successfully fetches and returns optimized route data', async () => {
    const mockResponseData: OptimizeResponse = {
      route_plan: {
        steps: [
          {
            shop_name: 'Aldi',
            items: ['tej'],
          },
        ],
      },
      total_cost: 1500,
    };

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponseData,
    });

    const result = await optimizeShoppingRoute('tej', 47.123, 19.456);

    expect(global.fetch).toHaveBeenCalledWith('http://localhost:8080/api/v1/optimize', {
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

  it('throws an error when the API returns a non-200 response with JSON error', async () => {
    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: false,
      status: 400,
      json: async () => ({ error: 'Invalid input', code: 400 }),
    });

    await expect(optimizeShoppingRoute('tej', 47.123, 19.456)).rejects.toThrow('Invalid input');
  });

  it('throws a generic error when the API returns a non-200 response without JSON', async () => {
    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: false,
      status: 500,
      json: async () => { throw new Error('Not JSON'); },
    });

    await expect(optimizeShoppingRoute('tej', 47.123, 19.456)).rejects.toThrow('HTTP error! status: 500');
  });
});
