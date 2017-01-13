# frozen_string_literal: true

class MaxValue < Grape::Validations::Base
  def validate_param!(attr_name, params)
    message = { params: [@scope.full_name(attr_name)], message: "must be small than #{@option}." }
    raise Grape::Exceptions::Validation, message if params[attr_name].to_i > @option
  end
end
