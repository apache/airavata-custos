export const decodeToken = (token: string | undefined) => {
  if (!token) return null;
  return JSON.parse(atob(token.split('.')[1]));
};
