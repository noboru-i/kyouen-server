# frozen_string_literal: true

module SharedParams
  extend Grape::API::Helpers

  params :pagination do
    optional :offset, type: Integer, default: 0
    optional :limit, type: Integer, default: 10, max_value: 100
  end
end

module KyouenServer
  module Entities
    class Stage < Grape::Entity
      expose :stage_no, documentation: { type: :integer, desc: 'Stage number.', required: true }
      expose :size, documentation: { type: :integer, desc: 'Size of stage.', required: true }
      expose :stage, documentation: { type: :string, desc: 'Describe stage state.', required: true }
      expose :creator, documentation: { type: :string, desc: 'Name of who created this.' }
    end
  end

  class API < Grape::API
    version 'v1', using: :header, vendor: 'kyouen'
    format :json

    require_relative '../validations/max_value'
    helpers SharedParams

    resource :stages do
      desc 'Return list of stages.', \
           entity: Entities::Stage,
           success: { code: 200, model: Entities::Stage }
      params do
        use :pagination
      end
      get '/' do
        list = KyouenPuzzle.fetch(params[:offset].to_i, params[:limit].to_i)
        present list, with: Entities::Stage
      end
    end

    add_swagger_documentation \
      info: {
        title: 'API documentation of kyouen app.'
      }
  end
end
