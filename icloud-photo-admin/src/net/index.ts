import axios from "axios";


export const login = async (account: string, password: string) => {
  const response = await axios.post(
    '/api/login', {
      account,
      password,
    });
  const {success} = response.data;
  return {success}
}

export const register = async (account: string, password: string) => {
  const response = await axios.post(
    '/api/register', {
      account,
      password,
    });
  const {success} = response.data;
  return {success}
}


export const addAccount = async (account: string, password: string, twoFactorCode: string) => {
  const response = await axios.post(
    '/api/addAccount', {
      account,
      password,
      twoFactorCode,
    });
  const {needsTwoFactor, success} = response.data;
  return {needsTwoFactor, success}
}

export const delAccount = async (account: string) => {
  const response = await axios.post(
    '/api/delAccount', {
      account,
    });
  const {success} = response.data;
  return {success}
}