interface ErrorResponse {
  error_code: number;
  error: string;
}

export const handleResponse = async (response: Response) => {
  if (!response.ok) {
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
    const result: ErrorResponse = await response.json();
    throw Error(result.error);
  }
  return response.json();
};

export const isWidthBelow = (width: number): boolean =>
  window.innerWidth < width;
