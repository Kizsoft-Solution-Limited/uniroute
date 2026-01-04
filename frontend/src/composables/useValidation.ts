import { ref, computed } from 'vue'
import * as yup from 'yup'
import { useField, useForm } from 'vee-validate'

/**
 * Composable for form validation using VeeValidate and Yup
 */
export function useValidation<T extends Record<string, any>>(
  schema: yup.ObjectSchema<T>,
  initialValues?: Partial<T>
) {
  const { handleSubmit, errors, values, setFieldValue, resetForm } = useForm<T>({
    validationSchema: schema,
    initialValues
  })

  const isValid = computed(() => Object.keys(errors.value).length === 0)

  const validate = async (): Promise<boolean> => {
    try {
      await schema.validate(values.value, { abortEarly: false })
      return true
    } catch (error) {
      return false
    }
  }

  return {
    handleSubmit,
    errors,
    values,
    isValid,
    validate,
    setFieldValue,
    resetForm
  }
}

/**
 * Composable for single field validation
 */
export function useFieldValidation<T = any>(
  name: string,
  schema: yup.Schema<T>,
  initialValue?: T
) {
  const { value, errorMessage, handleChange, handleBlur } = useField(name, schema, {
    initialValue
  })

  return {
    value,
    error: errorMessage,
    handleChange,
    handleBlur
  }
}

/**
 * Common validation schemas
 */
export const validationSchemas = {
  email: yup.string().email('Invalid email address').required('Email is required'),
  password: yup
    .string()
    .min(8, 'Password must be at least 8 characters')
    .matches(/[A-Z]/, 'Password must contain at least one uppercase letter')
    .matches(/[a-z]/, 'Password must contain at least one lowercase letter')
    .matches(/[0-9]/, 'Password must contain at least one number')
    .required('Password is required'),
  required: (message = 'This field is required') => yup.string().required(message),
  url: yup.string().url('Invalid URL').required('URL is required'),
  apiKey: yup.string().min(10, 'API key is too short').required('API key is required')
}

