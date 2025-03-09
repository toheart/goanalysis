import axios from 'axios';

// 创建axios实例
const instance = axios.create({
  baseURL: process.env.VUE_APP_API_URL || '',
  timeout: 30000, // 请求超时时间增加到30秒
  headers: {
    'Content-Type': 'application/json'
  }
});

// 请求拦截器
instance.interceptors.request.use(
  config => {
    console.log('发送请求:', {
      method: config.method,
      url: config.url,
      data: config.data
    });
    return config;
  },
  error => {
    console.error('请求错误:', error);
    return Promise.reject(error);
  }
);

// 响应拦截器
instance.interceptors.response.use(
  response => {
    console.log('收到响应:', {
      status: response.status,
      url: response.config.url,
      data: response.data
    });
    return response;
  },
  error => {
    console.error('响应错误:', error);
    if (error.response) {
      console.error('错误详情:', {
        status: error.response.status,
        data: error.response.data
      });
    } else if (error.request) {
      console.error('未收到响应:', error.request);
    }
    return Promise.reject(error);
  }
);

export default instance;