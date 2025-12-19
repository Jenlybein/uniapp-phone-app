import OpenAI from "openai";
// DeepSeek API 封装

// API配置 - 请替换为实际的API Key和请求地址
const API_CONFIG = {
  API_KEY: "24325ad4-1a64-4205-a311-68c91ede6a28",
  BASE_URL: "https://ark.cn-beijing.volces.com/api/v3",
  MODEL: "doubao-seed-1-6-251015",
};

const openai = new OpenAI({
  baseURL: API_CONFIG.BASE_URL,
  apiKey: API_CONFIG.API_KEY,
  dangerouslyAllowBrowser: true,
});

/**
 * 获取AI回复
 * @param {Object} params - 请求参数
 * @param {string} params.type - 内容类型：text或image
 * @param {string} params.content - 内容：文本或图片路径
 * @returns {Promise<string>} AI回复内容
 */
export const callDeepSeekApi = async (params) => {
  try {
    let requestData = {
      model: API_CONFIG.MODEL,
      messages: [],
    };

    // 构建对话内容
    if (params.type === "text") {
      requestData.messages.push({
        role: "user",
        content: params.content,
      });
    } else if (params.type === "image") {
      // 图片类型需要先进行Base64编码
      const base64Image = await getImageBase64(params.content);

      requestData.messages.push({
        role: "user",
        content: [
          {
            type: "text",
            text: "请根据图中问题解答",
          },
          {
            type: "image_url",
            image_url: {
              url: `data:image/jpeg;base64,${base64Image}`,
            },
          },
        ],
      });
    }

    // 发送请求
    const completion = await openai.chat.completions.create(requestData);

    // 校验响应整体结构
    if (!completion || !completion.choices || completion.choices.length === 0) {
      throw new Error("API响应格式异常：无有效回复内容");
    }

    // 获取第一个回复选项（DeepSeek通常只返回一个choice）
    const firstChoice = completion.choices[0];

    return completion.choices[0].message.content;
  } catch (error) {
    console.error("DeepSeek API调用错误:", error);
    throw error;
  }
};

/**
 * 将图片转换为Base64编码
 * @param {string} imagePath - 图片路径
 * @returns {Promise<string>} Base64编码的图片数据
 */
const getImageBase64 = (imagePath) => {
  return new Promise((resolve, reject) => {
    // 检查运行环境，适应不同平台
    if (typeof uni !== "undefined" && uni.getFileSystemManager) {
      // 小程序环境
      try {
        const fs = uni.getFileSystemManager();
        fs.readFile({
          filePath: imagePath,
          encoding: "base64",
          success: (res) => {
            resolve(res.data);
          },
          fail: (err) => {
            reject(new Error(`图片读取失败: ${err.errMsg}`));
          },
        });
      } catch (error) {
        reject(new Error(`小程序环境下图片处理失败: ${error.message}`));
      }
    } else if (typeof window !== "undefined") {
      // H5环境
      const img = new Image();
      img.crossOrigin = "anonymous"; // 解决跨域问题
      img.src = imagePath;

      img.onload = () => {
        const canvas = document.createElement("canvas");
        const ctx = canvas.getContext("2d");
        if (!ctx) {
          reject(new Error("无法获取Canvas上下文"));
          return;
        }

        // 设置Canvas尺寸与图片一致
        canvas.width = img.width;
        canvas.height = img.height;

        // 绘制图片到Canvas
        ctx.drawImage(img, 0, 0, img.width, img.height);

        try {
          // 将Canvas转换为Base64
          const base64 = canvas.toDataURL("image/jpeg").split(",")[1];
          resolve(base64);
        } catch (error) {
          reject(new Error(`Canvas转换Base64失败: ${error.message}`));
        }
      };

      img.onerror = (error) => {
        reject(new Error(`图片加载失败: ${error.message}`));
      };
    } else {
      // 其他环境，暂时不支持
      reject(new Error("当前环境不支持图片Base64转换"));
    }
  });
};
